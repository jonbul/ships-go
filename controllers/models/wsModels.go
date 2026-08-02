package models

type WsEvent struct {
	EventName string  `json:"eventName" bson:"eventName"`
	SocketId  string  `json:"socketId" bson:"socketId"`
	Bullet    Bullet  `json:"bullet" bson:"bullet"`
	X         float32 `json:"x" bson:"x"` // TODO REMOVE used in newBullet event
	Y         float32 `json:"y" bson:"y"` // TODO REMOVE used in newBullet event

	// TODO
	//hitData    playerHitData `json:"playerHitData" bson:"playerHitData"`
	//playerData playerData    `json:"playerData" bson:"playerData"`

	// TODO move to subclass playerHit
	BulletId     string  `json:"bulletId" bson:"bulletId"`         // TODO move to subclass playerHit
	PlayerId     string  `json:"playerId" bson:"playerId"`         // TODO move to subclass playerHit
	From         string  `json:"from" bson:"from"`                 // TODO move to subclass playerHit
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"` // TODO move to subclass playerHit
	// TODO move to subclass playerHit

	// TODO moveToSubClass player
	// x, y, socketId?, eventName?
	Credits      int     `json:"credits" bson:"credits"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	Scale        float32 `json:"scale" bson:"scale"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
	Name         string  `json:"name" bson:"name"`
	Life         float32 `json:"life" bson:"life"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`

	// TODO moveToSubClass player
}

/*
type PlayerData struct {
	X            int     `json:"x" bson:"x"`
	Y            int     `json:"y" bson:"y"`
	Credits      int     `json:"credits" bson:"credits"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	Scale        float32 `json:"scale" bson:"scale"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
	Name         string  `json:"name" bson:"name"`
	Life         float32 `json:"life" bson:"life"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`
}

type playerHitData struct {
	BulletId     string  `json:"bulletId" bson:"bulletId"`
	PlayerId     string  `json:"playerId" bson:"playerId"`
	From         string  `json:"from" bson:"from"`
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"`
}*/

type Bullet struct {
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
