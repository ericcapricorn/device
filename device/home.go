package device

type Home struct {
	hid       int64
	name      string
	createUid int64
	status    int8
}

func NewHome(hid int64, name string, uid int64, status int8) *Home {
	return &Home{hid: hid, name: name, createUid: uid, status: status}
}

func (this *Home) GetHid() int64 {
	return this.hid
}

func (this *Home) GetName() string {
	return this.name
}

func (this *Home) GetCreateUid() int64 {
	return this.createUid
}

func (this *Home) GetStatus() int8 {
	return this.status
}
