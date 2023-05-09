package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// sql.DB connection 풀을 감싸는 SnippetModel 유형을 정의합니다.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// 실행할 SQL 문을 작성합니다. 가독성을 위해 두 줄로 나누었습니다.
	// 가독성을 위해 (일반 큰따옴표 대신에
	// 큰따옴표로 묶은 이유입니다).
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// 임베디드 연결 풀에서 Exec() 메서드를 사용하여
	// 문을 실행합니다. 첫 번째 매개변수는 SQL 문이며, 그 뒤에 플레이스홀더 매개변수에 대한
	// 플레이스홀더 매개변수의 제목, 내용 및 만료 값입니다. 이
	// 메서드는 몇 가지 기본 정보를 포함하는 sql.Result 유형을 반환합니다.
	// 문이 실행되었을 때 어떤 일이 일어났는지에 대한 몇 가지 기본 정보가 포함된 쿼리 결과 유형을 반환합니다.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// 결과에서 LastInsertId() 메서드를 사용하여 새로 삽입된 레코드의 ID를 가져옵니다.
	// 코드조각 테이블에 새로 삽입된 레코드의 ID를 가져옵니다.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 반환된 ID의 유형이 int64이므로 반환하기 전에 int 유형으로 변환합니다.
	return int(id), nil
}

// 해당 ID를 기반으로 특정 스니펫이 반환됩니다.
func (m *SnippetModel) Get(id int) (*Snippet, error) {

	s := &Snippet{}
	err := m.DB.QueryRow(`SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

// 가장 최근에 생성된 10개 스니펫이 반환됩니다.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	// 실행할 SQL 문을 작성합니다.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// 연결 풀에서 Query() 메서드를 사용하여
	// SQL 문을 실행합니다. 그러면 쿼리 결과가 포함된 sql.Rows 결과 집합이 반환됩니다.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// sql.Rows 결과 집합이  항상 Latest() 메서드가 반환되기 전에 올바르게 닫히도록 합니다.
	// 이 지연문은 Query() 메서드에서 오류가 있는지 확인한 후 와야 합니다.
	// 그렇지 않으면, Query()가 오류를 반환하면 패닉 상태가 됩니다.
	// nil 결과 집합을 닫으려고 합니다.
	defer rows.Close()

	// 코드조각 구조체를 보관할 빈 슬라이스를 초기화합니다.
	snippets := []*Snippet{}
	// rows.Next를 사용하여 결과 집합의 행을 반복합니다. 이
	// 에 의해 동작할 첫 번째(그리고 이후의 각 행) 행을 준비합니다.
	// 모든 행에 대한 반복이 완료되면 결과 집합이 자동으로 닫히고
	//기본 데이터베이스 연결을 해제합니다.
	for rows.Next() {
		// 0이 된 새로운 코드조각 구조체에 대한 포인터를 생성합니다.
		s := &Snippet{}
		// 행의 각 필드에서 값을 행의 각 필드에 있는
		// 새 코드조각 객체로 복사합니다. 다시 말하지만, row.Scan()의 인수는 데이터를 복사하려는 위치에 대한 포인터여야 하며,
		// 인수의 수는 문에서 반환된 열의 수와 정확히 같아야 합니다.
		// 열의 수와 정확히 같아야 합니다.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		// 스니펫 조각에 추가합니다.
		snippets = append(snippets, s)
	}
	// rows.Next() 루프가 완료되면 rows.Err()를 호출하여 반복 중에 발생한 모든
	// 오류를 검색합니다. 이 함수를 호출하는 것이 중요합니다.
	// 완료되었다고 가정하지 마세요.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// 모든 것이 정상적으로 진행되었다면 코드조각 조각을 반환합니다.
	return snippets, nil
}
