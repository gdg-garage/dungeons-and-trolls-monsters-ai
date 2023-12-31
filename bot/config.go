package bot

type Config struct {
	Aggression   float32
	Preservation float32
	Support      float32

	Restlessness float32
	Randomness   float32
}

func NewConfig(algorithm string) Config {
	return Config{
		Aggression:   4,
		Preservation: 2,
		Support:      1.5,

		Restlessness: 1.2,
		Randomness:   0.03,
	}
}
