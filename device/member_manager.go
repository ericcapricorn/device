package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type MemberManager struct {
	store *DeviceStorage
}

func NewMemberManager(store *DeviceStorage) *MemberManager {
	return &MemberManager{store: store}
}

////////////////////////////////////////////////////////////////////////////////////
// public interface
////////////////////////////////////////////////////////////////////////////////////
func (this *MemberManager) Domain() string {
	return this.store.GetDomain()
}

// get member for check if the user has privelige
func (this *MemberManager) Get(hid, uid int64) (*Member, error) {
	common.CheckParam(this.store != nil)
	var member Member
	err := this.getMemberInfo(hid, uid, &member)
	if err != nil {
		if err == common.ErrEntryNotExist {
			return nil, nil
		} else {
			log.Warningf("get member info failed:domain[%d], hid[%d], uid[%d]", this.Domain(), hid, uid)
			return nil, err
		}
	}
	return &member, nil
}

// get all homeids belong to this member, if no hid return empty list
func (this *MemberManager) GetAllHomeIds(uid int64) ([]int64, error) {
	common.CheckParam(this.store != nil)
	list, err := this.getMemberAllHomes(uid)
	if err != nil {
		log.Warningf("get member all home ids failed:domain[%s], uid[%d], err[%v]",
			this.Domain(), uid, err)
		return nil, err
	}
	return list, nil
}

// add home owner
func (this *MemberManager) AddOwner(owner string, hid, uid int64) error {
	common.CheckParam(this.store != nil)
	if len(owner) <= 0 {
		log.Warningf("check owner name failed:domain[%s], owner[%s], uid[%d], hid[%d]",
			this.Domain(), owner, uid, hid)
		return common.ErrInvalidName
	}
	return this.addOwnerMember(owner, hid, uid)
}

// add home normal member
func (this *MemberManager) AddMember(member string, hid, uid int64) error {
	common.CheckParam(this.store != nil)
	if len(member) <= 0 {
		log.Warningf("check member name failed:domain[%s], member[%s], uid[%d], hid[%d]",
			this.Domain(), member, uid, hid)
		return common.ErrInvalidName
	}
	// step 1. get home info
	homeManager := NewHomeManager(this.store)
	home, err := homeManager.Get(hid)
	if err != nil {
		log.Warningf("get home failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return err
	} else if home == nil {
		log.Warningf("home not exist:domain[%s], hid[%d]", this.Domain(), hid)
		return common.ErrEntryNotExist
	} else if home.GetStatus() != ACTIVE {
		log.Warningf("check home status failed:domain[%s], hid[%d]", this.Domain(), hid)
		return common.ErrInvalidStatus
	} else if home.GetCreateUid() == uid {
		// do not return error
		log.Warningf("the user is creator:domain[%s], hid[%d], uid[%d]", this.Domain(), hid, home.createUid)
		return nil
	}
	// step 2. add member, if exist return error
	err = this.insertMemberInfo(member, hid, uid)
	if err != nil {
		log.Warningf("insert one member to home faileddomain[%s], hid[%d], uid[%d]", this.Domain(), hid, uid)
		return err
	}
	return nil
}

// del a member from a home
func (this *MemberManager) Delete(hid, uid int64) error {
	common.CheckParam(this.store != nil)
	homeManager := NewHomeManager(this.store)
	// step 1. check home info
	home, err := homeManager.Get(hid)
	if err != nil {
		log.Warningf("get home failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return err
	} else if home == nil {
		log.Warningf("home not exist:domain[%s], hid[%d]", this.Domain(), hid)
		return common.ErrEntryNotExist
	} else if home.GetStatus() != ACTIVE {
		log.Warningf("check home status failed:domain[%s], hid[%d]", this.Domain(), hid)
		return common.ErrInvalidStatus
	}
	// step 2. delete member, if not exist return succ
	err = this.deleteOneMember(hid, uid)
	if err != nil {
		log.Warningf("delete one member of home failed:domain[%s], hid[%d], err[%s]", this.Domain(), hid, err)
		return err
	}
	return nil
}

// get all members of the home, if (hid) not exist,
// return empty list, not nil + nil
func (this *MemberManager) GetAllMembers(hid int64) ([]Member, error) {
	common.CheckParam(this.store != nil)
	// if home not exist, return empty list
	list, err := this.getAllMembers(hid)
	if err != nil {
		log.Warningf("get one home all members failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return nil, err
	}
	return list, nil
}

// delete the home and delete all the members
func (this *MemberManager) DeleteAllMembers(hid int64) error {
	common.CheckParam(this.store != nil)
	err := this.deleteAllMembers(hid)
	if err != nil {
		log.Warningf("delete home all members failed:domain[%s], hid[%d], err[%s]",
			this.Domain(), hid, err)
		return err
	}
	return nil
}

// defrozen a member
func (this *MemberManager) Enable(hid, uid int64) error {
	return this.modifyMemberInfo(hid, uid, "status", 1)
}

// frozen a member
func (this *MemberManager) Disable(hid, uid int64) error {
	return this.modifyMemberInfo(hid, uid, "status", 0)
}

// modify member name in this home
func (this *MemberManager) ModifyName(hid, uid int64, name string) error {
	return this.modifyMemberInfo(hid, uid, "name", name)
}

////////////////////////////////////////////////////////////////////////////////////
// database related private interface
////////////////////////////////////////////////////////////////////////////////////
func (this *MemberManager) getMemberInfo(hid, uid int64, member *Member) error {
	SQL := fmt.Sprintf("SELECT uid, hid, name, type, status FROM %s_home_members WHERE uid = ? AND hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(uid, hid).Scan(&member.uid, &member.hid, &member.memberName, &member.memberType, &member.status)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrEntryNotExist
		} else {
			log.Warningf("query member info failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
			return err
		}
	}
	return nil
}

func (this *MemberManager) getMemberAllHomes(uid int64) ([]int64, error) {
	SQL := fmt.Sprintf("SELECT hid FROM %s_home_members WHERE uid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], uid[%d], err[%v]",
			this.Domain(), uid, err)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Errorf("query all homes failed:domain[%s], uid[%d], err[%v]",
			this.Domain(), uid, err)
		return nil, err
	}
	var hid int64
	list := make([]int64, 0)
	for rows.Next() {
		err = rows.Scan(&hid)
		if err != nil {
			log.Warningf("parse the result failed:domain[%s], uid[%d], err[%v]",
				this.Domain(), uid, err)
			return nil, err
		}
		list = append(list, hid)
	}
	return list, nil
}

// modify member info all column can be modified
func (this *MemberManager) modifyMemberInfo(hid, uid int64, key string, value interface{}) error {
	common.CheckParam(len(key) != 0)
	SQL := fmt.Sprintf("UPDATE %s_home_members SET %s = ? WHERE uid = ? AND hid = ?", this.Domain(), key)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, uid, hid)
	if err != nil {
		log.Errorf("execute update failed:sql[%s], err[%v]", SQL, err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Warningf("get affected rows failed:err[%v]", err)
		return err
	}
	if affect < 1 {
		log.Warningf("check affected rows failed:domain[%s], hid[%d], uid[%d], row[%d]", this.Domain(), hid, uid, affect)
		return common.ErrEntryNotExist
	} else if affect > 1 {
		log.Errorf("check affected rows failed:domain[%s], hid[%d], uid[%d], row[%d]", this.Domain(), hid, uid, affect)
		return common.ErrUnknown
	}
	return nil
}

func (this *MemberManager) addOwnerMember(owner string, hid, uid int64) error {
	SQL := fmt.Sprintf("INSERT INTO %s_home_members(uid, hid, type, name, status) VALUES(?,?,?,?,?)",
		this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], err[%s]", this.Domain(), err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uid, hid, MASTER, owner, ACTIVE)
	if err != nil {
		log.Warningf("insert owner failed:domain[%s], uid[%d], hid[%d]", this.Domain(), uid, hid)
		return err
	}
	return nil
}

func (this *MemberManager) getAllMembers(hid int64) ([]Member, error) {
	SQL := fmt.Sprintf("SELECT uid, hid, name, type, status FROM %s_home_members WHERE hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(hid)
	if err != nil {
		log.Errorf("query all members failed:domain[%s], hid[%d], err[%v]",
			this.Domain(), hid, err)
		return nil, err
	}
	defer rows.Close()
	var member Member
	list := make([]Member, 0)
	for rows.Next() {
		err := rows.Scan(&member.uid, &member.hid, &member.memberName, &member.memberType, &member.status)
		if err != nil {
			log.Errorf("parse the uid failed:domain[%s], hid[%d], err[%v]",
				this.Domain(), hid, err)
			return nil, err
		}
		list = append(list, member)
	}
	return list, nil
}

func (this *MemberManager) deleteOneMember(hid, uid int64) error {
	SQL := fmt.Sprintf("DELETE FROM %s_home_members WHERE uid = ? AND hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], hid[%d], err[%s]", this.Domain(), hid, err)
		return err
	}
	defer stmt.Close()
	// not check the affected rows
	_, err = stmt.Exec(uid, hid)
	if err != nil {
		log.Warningf("delete member failed:domain[%s], uid[%d], hid[%d]",
			this.Domain(), uid, hid)
		return err
	}
	return nil
}

func (this *MemberManager) deleteAllMembers(hid int64) error {
	SQL := fmt.Sprintf("DELETE FROM %s_home_members WHERE hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], hid[%d], err[%s]",
			this.Domain(), hid, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(hid)
	if err != nil {
		log.Warningf("delete all member failed:domain[%s], hid[%d]", this.Domain(), hid)
		return err
	}
	return nil
}

func (this *MemberManager) insertMemberInfo(member string, hid, uid int64) error {
	SQL := fmt.Sprintf("INSERT INTO %s_home_members(uid, hid, type, name, status) VALUES(?,?,?,?,?)",
		this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], err[%s]", this.Domain(), err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uid, hid, NORMAL, member, ACTIVE)
	if err != nil {
		log.Warningf("insert the member failed:domain[%s], name[%s], uid[%d], hid[%d]", this.Domain(), member, uid, hid)
		return err
	}
	return nil
}
