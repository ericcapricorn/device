package main

import (
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type DeviceService struct {
	zc.ZService
	home      *HomeManagerHandler
	member    *MemberManagerHandler
	dev       *DeviceManagerHandler
	warehouse *DeviceWarehouseHandler
	access    *DeviceAccessPointHandler
}

func (this *DeviceService) Validate() bool {
	return this.home != nil && this.member != nil && this.dev != nil &&
		this.warehouse != nil && this.access != nil
}

func NewDeviceService(database string, config *zc.ZServiceConfig) *DeviceService {
	const host string = "101.251.106.4:3306"
	const user string = "root"
	const password string = "123456"
	store := device.NewDeviceStorage(host, user, password, database)
	if store == nil {
		log.Fatalln("device storage init failed")
		return nil
	}
	home := NewHomeManagerHandler(device.NewHomeManager(store))
	member := NewMemberManagerHandler(device.NewMemberManager(store))
	dev := NewDeviceManagerHandler(device.NewDeviceManager(store), device.NewBindingManager(store))
	warehouse := NewDeviceWarehouseHandler(device.NewDeviceWarehouse(store))
	access := NewDeviceAccessPointHandler(device.NewAccessRouter(store))
	service := &DeviceService{home: home, member: member, dev: dev, warehouse: warehouse, access: access}
	if !service.Validate() {
		log.Fatalln("service init failed")
		return nil
	}
	service.Init("zc-dm", config)

	// device ctrl access point handler
	service.Handle("getapoint", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		access.handleGetAccessPoint(req, resp)
	}))

	// device warehouse handler
	service.Handle("registdevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		warehouse.handleRegistDevice(req, resp)
	}))
	service.Handle("getpublickey", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		warehouse.handleGetPublicKey(req, resp)
	}))

	// home manager handler
	service.Handle("listhomes", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		home.handleListHomes(req, resp)
	}))
	service.Handle("createhome", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		home.handleCreateHome(req, resp)
	}))
	service.Handle("modifyhome", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		home.handleModifyHome(req, resp)
	}))
	service.Handle("deletehome", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		home.handleDeleteHome(req, resp)
	}))
	service.Handle("frozenhome", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		home.handleFrozenHome(req, resp)
	}))

	// member manager handler
	service.Handle("listmembers", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		member.handleListMembers(req, resp)
	}))
	service.Handle("addmember", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		member.handleAddMember(req, resp)
	}))
	service.Handle("deletemember", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		member.handleDeleteMember(req, resp)
	}))
	service.Handle("modifymember", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		member.handleModifyMember(req, resp)
	}))
	service.Handle("frozenmember", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		member.handleFrozenMember(req, resp)
	}))

	// device manager handler
	service.Handle("listdevices", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleListDevices(req, resp)
	}))
	service.Handle("binddevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleBindDevice(req, resp)
	}))
	service.Handle("changedevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleChangeDevice(req, resp)
	}))
	service.Handle("deletedevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleDeleteDevice(req, resp)
	}))
	service.Handle("modifydevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleModifyDevice(req, resp)
	}))
	service.Handle("frozendevice", zc.ZServiceHandler(func(req *zc.ZMsg, resp *zc.ZMsg) {
		dev.handleFrozenDevice(req, resp)
	}))
	return service
}

func main() {
	var serverConfig = &zc.ZServiceConfig{Port: "5354"}
	server := NewDeviceService("device", serverConfig)
	if server == nil {
		log.Fatal("new device server failed, exit")
		return
	}
	// TODO defer close all the connections
	server.Start()
}
