package main

import (
	context "context"
	data "echsylon/fudpucker/data"
	message "echsylon/fudpucker/message"
	request "echsylon/fudpucker/request"
	signal "os/signal"
	syscall "syscall"

	"github.com/echsylon/go-log"
)

type Controller interface {
	SetupInfrastructure(apiServerPort int, messageServerPort int)
	StartApiServer()
}

type controller struct {
	mainContext      context.Context
	shutdownFunction context.CancelFunc
	properties       data.Preferences
	database         data.Database
	peers            data.PeerCache
	cache            data.MessageCache
	udp              message.UdpServer
	api              request.HttpServer
}

func NewController() Controller {
	ctxt, cncl := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &controller{
		mainContext:      ctxt,
		shutdownFunction: cncl,
	}
}

func (c *controller) SetupInfrastructure(apiServerPort int, messageServerPort int) {
	// Infrastructure
	c.properties = data.NewPreferences(apiServerPort, messageServerPort)
	c.database = data.NewDiskDatabase("./data/internal/database")
	c.peers = data.NewPeerCache()
	c.cache = data.NewMessageCache()
	c.udp = message.NewUdpServer(messageServerPort)
	c.api = request.NewHttpServer(c.mainContext, apiServerPort)

	// Adapters and handlers
	deviceProvider := data.NewGetDeviceDataAdapter(c.database.Get)
	deviceOwnerProvider := data.NewGetDeviceOwnerAttributeAdapter(c.database.Get)
	stateVersionProvider := data.NewGetStateVersionAttributeAdapter(c.database.Get)
	deviceIdsProvider := data.NewGetDeviceIdsDataAdapter(c.database.Get)
	devicePersister := data.NewCreateDeviceDataAdapter(c.database.Set)
	statePersister := data.NewPatchStateDataAdapter(c.database.Set)

	patchMessageProvider := message.NewPatchMessageProvider(c.properties.GetHostId)
	patchMessageReader := message.NewPatchMessageReader()
	syncMessageProvider := message.NewSyncMessageProvider(c.properties.GetHostId)
	syncMessageReader := message.NewSyncMessageReader()
	peerMessageProvider := message.NewPeerMessageProvider(c.properties.GetHostId)
	peerMessageReader := message.NewPeerMessageReader()
	hailMessageProvider := message.NewHailMessageProvider(c.properties.GetHostId)
	farewellMessageProvider := message.NewFarewellMessageProvider(c.properties.GetHostId)
	sendMessageHandler := message.NewSendMessageHandler(c.udp.Send, c.cache.Hold)

	// Usecases
	createDeviceUseCase := data.NewCreateDeviceUseCase(c.properties.GetHostId, devicePersister)
	checkIfOwnerUseCase := data.NewCheckIfOwnerUseCase(c.properties.GetHostId, deviceOwnerProvider)
	checkIfNewerUseCase := data.NewCheckIfNewerUseCase(stateVersionProvider)
	composeInfoUseCase := data.NewGetHostInfoUseCase(
		c.properties.GetHostId,
		c.properties.GetLocalAddress,
		c.properties.GetBroadcastAddress,
	)
	getRandomPeersUseCase := data.NewRandomSafePeersForMessageUseCase(
		c.properties.GetHostId,
		c.properties.GetBroadcastAddress,
		c.peers.GetAllPeers,
		c.cache.ContainsMessageForPeer,
	)
	patchStateUseCase := data.NewPatchStateUseCase(
		checkIfOwnerUseCase,
		deviceProvider,
		statePersister,
		syncMessageProvider,
		patchMessageProvider,
		getRandomPeersUseCase,
		sendMessageHandler,
	)
	updateStateUseCase := message.NewSaveStateUseCase(
		patchMessageReader,
		checkIfOwnerUseCase,
		deviceProvider,
		statePersister,
		syncMessageProvider,
		getRandomPeersUseCase,
		sendMessageHandler,
	)
	updateDeviceUseCase := message.NewSaveDeviceUseCase(
		checkIfOwnerUseCase,
		checkIfNewerUseCase,
		syncMessageReader,
		devicePersister,
		deviceProvider,
		getRandomPeersUseCase,
		syncMessageProvider,
		sendMessageHandler,
	)
	updatePeerUseCase := message.NewSavePeerUseCase(
		peerMessageReader,
		c.peers.AddPeer,
	)
	deletePeerUseCase := message.NewDeletePeerUseCase(
		c.peers.RemovePeer,
	)
	saluteOnHailUseCase := message.NewSaluteOnHailUseCase(
		deviceIdsProvider,
		deviceProvider,
		c.peers.GetAllPeers,
		c.peers.AddPeer,
		syncMessageProvider,
		peerMessageProvider,
		sendMessageHandler,
	)
	sendHailMessage := message.NewSendHailCommandUseCase(
		getRandomPeersUseCase,
		hailMessageProvider,
		deviceIdsProvider,
		deviceProvider,
		syncMessageProvider,
		sendMessageHandler,
	)
	sendFarewellMessage := message.NewSendFarewellEventUseCase(
		getRandomPeersUseCase,
		farewellMessageProvider,
		sendMessageHandler,
	)

	receivedMessageHandler := message.NewReceiveMessageHandler(
		saluteOnHailUseCase,
		updateStateUseCase,
		updateDeviceUseCase,
		updatePeerUseCase,
		deletePeerUseCase,
		c.cache.ContainsMessage,
		c.cache.Hold,
	)

	// Request components
	renderApiDocUseCase := request.NewGetApiDocUseCase()
	shutdownHandler := request.NewShutdownUseCase(c.peers.Reset, c.cache.Reset, c.shutdownFunction)
	deleteDeviceUseCase := request.NewDeleteDeviceUseCase(checkIfOwnerUseCase, c.database.Delete)
	getApiHandler := request.NewGetApiRequestHandler(renderApiDocUseCase)
	getInfoHandler := request.NewGetHostInfoRequestHandler(composeInfoUseCase)
	getDeviceIdsHandler := request.NewGetDeviceIdsRequestHandler(deviceIdsProvider)
	getDeviceHandler := request.NewGetDeviceRequestHandler(deviceProvider)
	patchStateHandler := request.NewPatchStateRequestHandler(patchStateUseCase)
	createDeviceHandler := request.NewCreateDeviceRequestHandler(createDeviceUseCase)
	deleteDeviceHandler := request.NewDeleteDeviceRequestHandler(deleteDeviceUseCase)
	getPeersRequestHandler := request.NewGetPeersRequestHandler(c.peers.GetAllPeers)
	addPeerRequestHandler := request.NewAddPeerRequestHandler(c.peers.AddPeer)
	shutdownRequestHandler := request.NewShutdownRequestHandler(shutdownHandler)
	joinNetworkRequestHandler := request.NewJoinNetworkRequestHandler(func() error {
		c.udp.Observe(receivedMessageHandler)
		sendHailMessage()
		return nil
	})
	leaveNetworkRequestHandler := request.NewLeaveNetworkRequestHandler(func() error {
		sendFarewellMessage()
		c.udp.Stop()
		return nil
	})

	c.api.Handle("GET /{$}", getApiHandler)
	c.api.Handle("GET /info", getInfoHandler)
	c.api.Handle("GET /device", getDeviceIdsHandler)
	c.api.Handle("GET /device/{id}", getDeviceHandler)
	c.api.Handle("PATCH /device/{id}", patchStateHandler)
	c.api.Handle("POST /device", createDeviceHandler)
	c.api.Handle("DELETE /device/{id}", deleteDeviceHandler)
	c.api.Handle("GET /peer", getPeersRequestHandler)
	c.api.Handle("POST /peer", addPeerRequestHandler)
	c.api.Handle("POST /network", joinNetworkRequestHandler)
	c.api.Handle("DELETE /network", leaveNetworkRequestHandler)
	c.api.Handle("POST /shutdown", shutdownRequestHandler)
}

func (c *controller) StartApiServer() {
	defer c.api.Stop()
	defer c.udp.Stop()
	go c.api.Serve()

	if c.mainContext.Err() == nil {
		log.Information("API Server started successfully")
	}

	// Wait for main context to be cancelled. This is done either by
	// terminating the process (i.e. "Ctrl+C") or by sending a POST
	// request to the /shutdown endpoin.
	<-c.mainContext.Done()
}
