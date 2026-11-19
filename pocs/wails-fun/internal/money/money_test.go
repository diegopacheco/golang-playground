package money

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name string
		in   Cents
		want string
	}{
		{"cents are padded so $1.05 is not shown as $1.5", 105, "$1.05"},
		{"zero is printed with cents", 0, "$0.00"},
		{"negative amounts keep the sign in front", -250, "-$2.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}
