package agent

import "testing"

func TestIsRiskyTodo(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{text: "Обновить README", want: false},
		{text: "drop table posts", want: true},
		{text: "rm -rf /tmp/cache", want: true},
		{text: "docker compose down -v", want: true},
		{text: "delete from posts where id=1", want: true},
	}

	for _, tc := range tests {
		got := isRiskyTodo(tc.text)
		if got != tc.want {
			t.Fatalf("isRiskyTodo(%q)=%v want %v", tc.text, got, tc.want)
		}
	}
}
