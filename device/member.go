package device

type Member struct {
	uid        int64
	hid        int64
	status     int8
	memberType int8
	memberName string
}

func NewMember(uid int64, hid int64, name string, memberType, status int8) *Member {
	return &Member{uid: uid, hid: hid, memberName: name, status: status, memberType: memberType}
}

func (this *Member) GetUid() int64 {
	return this.uid
}

func (this *Member) GetHid() int64 {
	return this.hid
}

func (this *Member) GetStatus() int8 {
	return this.status
}

func (this *Member) GetMemberName() string {
	return this.memberName
}

func (this *Member) GetMemberType() int8 {
	return this.memberType
}
