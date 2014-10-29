package main

import (
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type MemberManagerHandler struct {
	member *device.MemberManager
}

func NewMemberManagerHandler(member *device.MemberManager) *MemberManagerHandler {
	if member == nil {
		return nil
	}
	return &MemberManagerHandler{member: member}
}

////////////////////////////////////////////////////////////////////////////////////////////
/// MEMBER MANAGER
////////////////////////////////////////////////////////////////////////////////////////////
// list all members
func (this *MemberManagerHandler) handleListMembers(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	list, err := this.member.GetAllMembers(domain, hid)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("list all members failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return
	}
	for _, member := range list {
		resp.AddObject("homes", zc.ZObject{"hid": member.GetHid(), "id": member.GetUid(), "name": member.GetMemberName()})
	}
	log.Warningf("list all members succ:domain[%s], hid[%d], count[%d]", domain, hid, len(list))
	resp.SetAck()
}

// add member
func (this *MemberManagerHandler) handleAddMember(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	uid := req.GetInt("uid")
	name := req.GetString("uname")
	err := this.member.AddMember(domain, hid, uid, name)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("add member failed:domain[%s], hid[%d], uid[%d], uname[%s], err[%v]", domain, hid, uid, name, err)
		return
	}
	log.Infof("add member succ:domain[%s], hid[%d], uid[%d], uname[%s]", domain, hid, uid, name)
	resp.SetAck()
}

// delete member
func (this *MemberManagerHandler) handleDeleteMember(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	uid := req.GetInt("uid")
	err := this.member.Delete(domain, hid, uid)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("delete member failed:domain[%s], hid[%d], uid[%d], err[%v]", domain, hid, uid, err)
		return
	}
	log.Infof("delete member succ:domain[%s], hid[%d], uid[%d]", domain, hid, uid)
	resp.SetAck()
}

// modify member name
func (this *MemberManagerHandler) handleModifyMember(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	uid := req.GetInt("uid")
	name := req.GetString("uname")
	err := this.member.ModifyName(domain, hid, uid, name)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("modify member name failed:domain[%s], hid[%d], uid[%d], uname[%s], err[%v]", domain, hid, uid, name, err)
		return
	}
	log.Infof("modify member name succ:domain[%s], hid[%d], uid[%d], uname[%s]", domain, hid, uid, name)
	resp.SetAck()
}

// frozen/defrozen member
func (this *MemberManagerHandler) handleFrozenMember(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	uid := req.GetInt("uid")
	frozen := req.GetBool("frozen")
	var err error
	if frozen {
		err = this.member.Disable(domain, hid, uid)
	} else {
		err = this.member.Enable(domain, hid, uid)
	}
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("frozen/defrozen member failed:domain[%s], hid[%d], uid[%d], frozen[%t], err[%v]", domain, hid, uid, frozen, err)
		return
	}
	log.Infof("frozen/defrozen member succ:domain[%s], hid[%d], uid[%d], frozen[%t]", domain, hid, uid, frozen)
	resp.SetAck()
}
