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

func TestDetectTaskKind(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "сделать dbcheck", want: "db_check"},
		{text: "показать список таблиц", want: "tables_check"},
		{text: "проверить count(*) from posts", want: "count_posts"},
		{text: "прогнать make test", want: "health_check"},
		{text: "обновить документацию", want: "todo_item"},
	}
	for _, tc := range tests {
		if got := detectTaskKind(tc.text); got != tc.want {
			t.Fatalf("detectTaskKind(%q)=%q want %q", tc.text, got, tc.want)
		}
	}
}
