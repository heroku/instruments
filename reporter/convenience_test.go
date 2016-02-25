package reporter

/*
func TestGetOrRegisterInstrument(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", instruments.NewRate())
	i := r.Register("foo", instruments.NewGauge(0))
	if _, ok := i.(*instruments.Rate); !ok {
		t.Fatal("wrong instrument type")
	}
	registered := r.Instruments()
	if len(registered) != 1 {
		t.Fatal("registry should only have one instruments registered")
	}
	i, p := registered["foo"]
	if !p {
		t.Fatal("instrument not found")
	}
	if _, ok := i.(*instruments.Rate); !ok {
		t.Fatal("wrong instrument type")
	}
}
*/
