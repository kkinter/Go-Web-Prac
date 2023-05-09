package main

import (
	"testing"
	"time"

	"snippetbox.wook.net/internal/assert"
)

func TestHumanDate(t *testing.T) {

	// 테스트 케이스 이름, humanDate() 함수에 대한 입력(tm 필드),
	// 예상 출력(want 필드)을 포함하는 익명 구조체의 슬라이스를 만듭니다.
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2022 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2022, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Mar 2022 at 09:15",
		},
	}

	for _, tt := range tests {
		// 새로운 assert.Equal() 헬퍼를 사용하여 예상 값과 실제 값을 비교합니다.
		t.Run(tt.name, func(t *testing.T) {
			hd := humaDate(tt.tm)

			assert.Equal(t, hd, tt.want)
		})
	}
}
