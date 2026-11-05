package event

func Site() []string {
	return append(Shop(), Plan()...)
}
