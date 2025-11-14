package subjects

type Subject struct {
	Name string `db:"name"`
	ID   int64  `db:"id"`
}
type Subjects struct {
	Data []Subject
}
