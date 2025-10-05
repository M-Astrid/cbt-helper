package entity

type Emotion struct {
	ID   int64
	Name string
}

type SMEREmotion struct {
	EmotionID int64
	SMERID    int64
	Scale     int
}
