package main

import (
	"context"
	"fmt"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/loopfz/gadgeto/tonic/utils/jujerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/api"
	"github.com/scottrmalley/p2p-file-challenge/config"
	"github.com/scottrmalley/p2p-file-challenge/networking"
	"github.com/scottrmalley/p2p-file-challenge/protocol"
	"github.com/scottrmalley/p2p-file-challenge/repository"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = 5 * time.Minute

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "p2p-file-challenge"

// mustResolve is just a small helper function to keep the main function clean
func mustResolve[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// create a new libp2p Host that listens on a random TCP port
	node := mustResolve(
		libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")),
	)

	fmt.Println("Listen addresses:", node.Addrs())

	// create a new PubSub service using the GossipSub router
	ps := mustResolve(pubsub.NewGossipSub(ctx, node))

	// setup local mDNS discovery
	if err := setupDiscovery(node); err != nil {
		panic(err)
	}

	// initialize the set announcement subscription
	connection := networking.NewConnection(
		ps,
		node.ID(),
	)
	setAnnouncement := mustResolve(
		networking.NewSetAnnouncement(
			rootLogger.With().Str("ctx", "set-announcement").Logger(),
			connection,
		),
	)

	repo := repository.NewFiles(
		rootLogger.With().Str("ctx", "file-repo").Logger(),
		mustResolve(gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})),
	)

	if err := repo.Migrate(); err != nil {
		panic(err)
	}

	processor := protocol.NewProcessor(
		repo,
	)

	service := api.NewService(
		rootLogger.With().Str("ctx", "api-service").Logger(),
		protocol.NewBroadcaster(
			rootLogger.With().Str("ctx", "broadcaster").Logger(),
			connection,
			setAnnouncement,
		),
		repo,
		processor,
	)

	controller := api.NewController(
		rootLogger.With().Str("ctx", "api-controller").Logger(),
		service,
	)

	// stream new file sets to database
	streamer := protocol.NewStreamer(
		rootLogger.With().Str("ctx", "streamer").Logger(),
		processor,
		connection,
	)

	router := defaultGinInit()
	if err := controller.RegisterRoutes(router.Group("/api")); err != nil {
		panic(err)
	}

	group, groupCtx := errgroup.WithContext(ctx)
	env := config.ParseHttpEnv("SVC")
	group.Go(
		func() error {
			rootLogger.Info().Int("port", env.Port).Msg("starting http server")
			if err := manners.ListenAndServe(fmt.Sprintf(":%d", env.Port), router); err != nil {
				rootLogger.Error().Err(err).Msg("error starting server")
				return err
			}
			return nil
		},
	)
	group.Go(
		func() error {
			<-groupCtx.Done()
			rootLogger.Info().Msg("context canceled: shutting down http server")
			if ok := manners.Close(); !ok {
				return errors.New("failed to close http server")
			}
			return nil
		},
	)

	group.Go(streamer.WatchNew(groupCtx, setAnnouncement.Read(groupCtx)))

	if err := group.Wait(); err != nil {
		rootLogger.Fatal().Err(err).Msg("error in main")
	}

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
func defaultGinInit() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	tonic.SetErrorHook(jujerr.ErrHook)
	return router
}