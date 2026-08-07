package models

type WsEvent struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
}

type PlayerData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	Credits      int     `json:"credits" bson:"credits"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Life         float32 `json:"life" bson:"life"`
	Name         string  `json:"name" bson:"name"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Scale        float32 `json:"scale" bson:"scale"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	X            float32 `json:"x" bson:"x"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`
	Y            float32 `json:"y" bson:"y"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
}

type PlayerHitData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"`
	BulletId     string  `json:"bulletId" bson:"bulletId"`
	From         string  `json:"from" bson:"from"`
	PlayerId     string  `json:"playerId" bson:"playerId"`
	X            float32 `json:"x" bson:"x"`
	Y            float32 `json:"y" bson:"y"`
}

type BulletData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	Angle         float32 `json:"angle" bson:"angle"`
	BulletCharge  float32 `json:"bulletCharge" bson:"bulletCharge"`
	ExpY          float32 `json:"expY" bson:"expY"`
	ExpX          float32 `json:"expX" bson:"expX"`
	Id            string  `json:"id" bson:"id"`
	MoveX         float32 `json:"moveX" bson:"moveX"`
	MoveY         float32 `json:"moveY" bson:"moveY"`
	Rotation      float32 `json:"rotation" bson:"rotation"`
	ShootingSpeed float32 `json:"shootingSpeed" bson:"shootingSpeed"`
	X             float32 `json:"x" bson:"x"`
	Y             float32 `json:"y" bson:"y"`
}
