package main

import (
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type HomeManagerHandler struct {
	home *device.HomeManager
}

func NewHomeManagerHandler(home *device.HomeManager) *HomeManagerHandler {
	if home == nil {
		return nil
	}
	return &HomeManagerHandler{home: home}
}

////////////////////////////////////////////////////////////////////////////////////////////
/// HOME MANAGER
////////////////////////////////////////////////////////////////////////////////////////////
// list all homes
func (this *HomeManagerHandler) handleListHomes(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	uid := req.GetInt("uid")
	list, err := this.home.GetAllHome(domain, uid)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("list all home failed:domain[%s], uid[%d], err[%v]", domain, uid, err)
		return
	}
	for _, home := range list {
		resp.AddObject("homes", zc.ZObject{"id": home.GetHid(), "name": home.GetName()})
	}
	log.Warningf("list all home succ:domain[%s], uid[%d], count[%d]", domain, uid, len(list))
	resp.SetAck()
}

// create home
func (this *HomeManagerHandler) handleCreateHome(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	uid := req.GetInt("uid")
	name := req.GetString("hname")
	if len(name) == 0 {
		name = "Home"
	}
	err := this.home.Create(domain, uid, name)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("create home failed:domain[%s], uid[%d], name[%s], err[%v]", domain, uid, name, err)
		return
	}
	log.Infof("create home succ:domain[%s], uid[%d], name[%s]", domain, uid, name)
	resp.SetAck()
}

// modify home
func (this *HomeManagerHandler) handleModifyHome(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	name := req.GetString("hname")
	err := this.home.ModifyName(domain, hid, name)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("modify home name failed:domain[%s], hid[%d], name[%s], err[%v]", domain, hid, name, err)
		return
	}
	log.Infof("modify home name succ:domain[%s], hid[%d], name[%s]", domain, hid, name)
	resp.SetAck()
}

// delete home
func (this *HomeManagerHandler) handleDeleteHome(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	err := this.home.Delete(domain, hid)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("delete home failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return
	}
	log.Infof("delete home succ:domain[%s], hid[%d]", domain, hid)
	resp.SetAck()
}

// frozen/defrozen home
func (this *HomeManagerHandler) handleFrozenHome(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	frozen := req.GetBool("frozen")
	var err error
	if frozen {
		err = this.home.Disable(domain, hid)
	} else {
		err = this.home.Enable(domain, hid)
	}
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("frozen/defrozen home failed:domain[%s], hid[%d], frozen[%t], err[%v]", domain, hid, frozen, err)
		return
	}
	log.Infof("frozen/defrozen home succ:domain[%s], hid[%d], frozen[%t]", domain, hid, frozen)
	resp.SetAck()
}
