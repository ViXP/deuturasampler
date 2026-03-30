package factories

type FileFactorable interface {
	Create(rowsData [][]byte) []byte
}
