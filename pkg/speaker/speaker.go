package speaker

type Engine interface {
	Say(text string) (SpeechStream, error)
}
