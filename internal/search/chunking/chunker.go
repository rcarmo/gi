package chunking

// Chunk is the low-level chunk representation emitted by chunkers.
type Chunk struct {
	ChunkIndex    int
	StartByte     int
	EndByte       int
	StartLine     int
	EndLine       int
	TokenEstimate int
	Heading       string
	Content       string
}

// Chunker splits one file into stable searchable chunks.
type Chunker interface {
	Version() string
	Chunk(path string, data []byte) ([]Chunk, error)
}
