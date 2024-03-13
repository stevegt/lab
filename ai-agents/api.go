package agents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fabiustech/openai"
	fabius_models "github.com/fabiustech/openai/models"
	oai "github.com/sashabaranov/go-openai"
	. "github.com/stevegt/goadapt"
	"github.com/stevegt/semver"
	"github.com/tiktoken-go/tokenizer"
)

		"gpt-4":         {"", 8192, oai.GPT4, false},
// findModel returns the model name and model_t given a model name.
// if the given model name is empty, then use DefaultModel.
func (models *Models) findModel(model string) (name string, m *Model, err error) {
	if model == "" {
		model = DefaultModel
	}
	m, ok := models.Available[model]
	if !ok {
		err = fmt.Errorf("model %q not found", model)
		return
	}
	name = model
	return
}

// Document is a single document in a document repository.
type Document struct {
	// XXX deprecated because we weren't precise about what it meant.
	Path string
	// The path to the document file, relative to g.Root
	RelPath string
}

// AbsPath returns the absolute path of a document.
func (g *Grokker) AbsPath(doc *Document) string {
	return filepath.Join(g.Root, doc.RelPath)
}

// Chunk is a single chunk of text from a document.
type Chunk struct {
	// The document that this chunk is from.
	Document *Document
	// The offset of the chunk in the document.
	Offset int
	// The length of the chunk in the document.
	Length int
	// sha256 hash of the text of the chunk.
	Hash string
	// The text of the chunk.  This is not stored in the db.
	text string
	// The embedding of the chunk.
	Embedding []float64
	// The grokker that this chunk belongs to.
	// g *Grokker
	// true if needs to be garbage collected
	stale bool
}

// NewChunk creates a new chunk given an offset, length, and text. It
// computes the sha256 hash of the text if doc is not nil.  It does
// not compute the embedding or add the chunk to the db.
func NewChunk(doc *Document, offset, length int, text string) (c *Chunk) {
	var prefixedText string
	var hashStr string
	if doc != nil {
		prefixedText = fmt.Sprintf("from %s:\n%s\n", doc.RelPath, text)
		hash := sha256.Sum256([]byte(prefixedText))
		hashStr = hex.EncodeToString(hash[:])
	}
	c = &Chunk{
		// g:        g,
		Document: doc,
		Offset:   offset,
		Length:   length,
		Hash:     hashStr,
		text:     text,
	}
	Debug("NewChunk: %#v", c)
	return
}

// ChunkText returns the text of a chunk.
func (g *Grokker) ChunkText(c *Chunk, withHeader bool) (text string, err error) {
	Debug("ChunkText(%#v)", c)
	if c.Document == nil {
		Assert(c.text != "", "ChunkText: c.Document == nil && c.text == \"\"")
		text = c.text
	} else {
		// read the chunk from the document
		var buf []byte
		buf, err = ioutil.ReadFile(g.AbsPath(c.Document))
		if os.IsNotExist(err) {
			// document has been removed; don't remove it from the
			// database, but don't return any text either.  The
			// document might be on a different branch in e.g. git.
			err = nil
			return
		}
		Ck(err)
		start := c.Offset
		stop := c.Offset + c.Length
		if stop > len(buf) {
			stop = len(buf)
		}
		if start < len(buf) {
			text = string(buf[c.Offset:stop])
		}
		if withHeader {
			text = fmt.Sprintf("from %s:\n%s\n", c.Document.RelPath, text)
		}
	}
	return
}

type Grokker struct {
	embeddingClient *openai.Client
	chatClient      *oai.Client
	// The grokker version number this db was last updated with.
	Version string
	// The absolute path of the root directory of the document
	// repository.  This is passed in from cli based on where we
	// found the db.
	Root string
	// The list of documents in the database.
	Documents []*Document
	// The list of chunks in the database.
	Chunks []*Chunk
	// model specs
	models              *Models
	Model               string
	oaiModel            string
	tokenizer           tokenizer.Codec
	tokenLimit          int
	embeddingTokenLimit int
	grokpath            string
	// XXX use a real tokenizer and replace maxChunkLen with tokenLimit.
	// tokenLimit int
	// maxChunkLen          int
	// maxEmbeddingChunkLen int
	// the path to the grokker db
}

// Init creates a Grokker database in the given root directory.
func Init(rootdir, model string) (g *Grokker, err error) {
	defer Return(&err)
	g, err = InitNamed(rootdir, ".grok", model)
	return
}

// InitNamed creates a named Grokker database in the given root directory.
func InitNamed(rootdir, name, model string) (g *Grokker, err error) {
	// ensure rootdir is absolute and exists
	rootdir, err = filepath.Abs(rootdir)
	Ck(err)
	_, err = os.Stat(rootdir)
	Ck(err)
	// create the db
	g = &Grokker{
		Root:    rootdir,
		Version: version,
	}
	// initialize other bits
	err = g.setup(model)
	Ck(err)
	// ensure there is no existing db
	g.grokpath = filepath.Join(rootdir, name)
	_, err = os.Stat(g.grokpath)
	if err == nil {
		err = fmt.Errorf("db already exists at %q", g.grokpath)
		return
	}
	// save the db
	fh, err := os.Create(g.grokpath)
	Ck(err)
	err = g.Save()
	Ck(err)
	fh.Close()
	return
}

// Mtime returns the last modified time of the Grokker database.
func (g *Grokker) Mtime() (timestamp time.Time, err error) {
	defer Return(&err)
	fi, err := os.Stat(g.grokpath)
	Ck(err)
	timestamp = fi.ModTime()
	return
}

// Load loads a Grokker database from the current or any parent directory.
func Load() (g *Grokker, migrated bool, oldver, newver string, err error) {
	defer Return(&err)

	// find the .grok file in the current or any parent directory
	grokfnbase := ".grok"
	grokpath := ""
	for level := 0; level < 99; level++ {
		path := strings.Repeat("../", level) + grokfnbase
		if _, err := os.Stat(path); err == nil {
			grokpath = path
			break
		}
	}
	g, migrated, oldver, newver, err = LoadFrom(grokpath)
	Ck(err)
	return
}

// LoadFrom loads a Grokker database from a given path.
// XXX replace the json db with a kv store, store vectors as binary
// floating point values.
func LoadFrom(grokpath string) (g *Grokker, migrated bool, oldver, newver string, err error) {
	g = &Grokker{}
	g.grokpath = grokpath
	// load the db
	fh, err := os.Open(g.grokpath)
	Ck(err)
	buf, err := ioutil.ReadAll(fh)
	Ck(err)
	err = json.Unmarshal(buf, g)
	Ck(err)
	// set the root directory, overriding whatever was in the db
	// - this is necessary because the db might have been moved
	g.Root, err = filepath.Abs(filepath.Dir(g.grokpath))
	Ck(err)

	migrated, oldver, newver, err = g.migrate()
	Ck(err)

	err = g.setup(g.Model)
	Ck(err)
	return
}

// CodeVersion returns the version of grokker.
func (g *Grokker) CodeVersion() string {
	return version
}

// DBVersion returns the version of the grokker database.
func (g *Grokker) DBVersion() string {
	return g.Version
}

// migrate migrates the current Grokker database from an older version
// to the current version.
func (g *Grokker) migrate() (migrated bool, was, now string, err error) {
	defer Return(&err)

	was = g.Version
	now = g.Version

	// set default version
	if g.Version == "" {
		g.Version = "0.1.0"
	}

	// loop until migrations are done
	for {

		// check if migration is necessary
		var dbver, codever *semver.Version
		dbver, err = semver.Parse([]byte(g.Version))
		Ck(err)
		codever, err = semver.Parse([]byte(version))
		Ck(err)
		if semver.Cmp(dbver, codever) == 0 {
			// no migration necessary
			break
		}

		// see if db is newer version than code
		if semver.Cmp(dbver, codever) > 0 {
			// db is newer than code
			err = fmt.Errorf("grokker db is version %s, but you're running version %s -- upgrade grokker", g.Version, version)
			return
		}

		Fpf(os.Stderr, "migrating from %s to %s\n", g.Version, version)

		// if we get here, then dbver < codever
		_, minor, patch := semver.Upgrade(dbver, codever)
		Assert(patch, "patch should be true: %s -> %s", dbver, codever)

		// figure out what kind of migration we need to do
		if minor {
			// minor version changed; db migration necessary
			err = g.migrateDb()
			Ck(err)
		} else {
			// only patch version changed; a patch version change is
			// just a code change, so just update the version number
			// in the db
			g.Version = version
		}

		migrated = true
	}

	now = g.Version

	return
}

// Backup backs up the Grokker database to a time-stamped backup and
// returns the path.
func (g *Grokker) Backup() (backpath string, err error) {
	defer Return(&err)
	Assert(g.grokpath != "", "g.grokpath is empty")
	tmpdir := os.TempDir()
	deslashed := strings.Replace(g.grokpath, "/", "-", -1)
	backpath = fmt.Sprintf("%s/grokker-backup-%s%s", tmpdir, time.Now().Format("20060102-150405"), deslashed)
	err = copyFile(g.grokpath, backpath)
	Ck(err, "failed to backup %q to %q", g.grokpath, backpath)
	return
}

// setup the model and oai clients.
// This function needs to be idempotent because it might be called multiple
// times during the lifetime of a Grokker object.
func (g *Grokker) setup(model string) (err error) {
	defer Return(&err)
	err = g.initModel(model)
	Ck(err)
	g.initClients()
	// initialize the tokenizer
	g.tokenizer, err = tokenizer.Get(tokenizer.Cl100kBase)
	Ck(err)
	return
}

// initClients initializes the OpenAI clients.
// This function needs to be idempotent because it might be called multiple
// times during the lifetime of a Grokker object.
func (g *Grokker) initClients() {
	authtoken := os.Getenv("OPENAI_API_KEY")
	g.embeddingClient = openai.NewClient(authtoken)
	g.chatClient = oai.NewClient(authtoken)
	return
}

// initModel initializes the model for a new or reloaded Grokker database.
// This function needs to be idempotent because it might be called multiple
// times during the lifetime of a Grokker object.
func (g *Grokker) initModel(model string) (err error) {
	defer Return(&err)
	Assert(g.Root != "", "root directory not set")
	g.models = NewModels()
	model, m, err := g.models.findModel(model)
	Ck(err)
	m.active = true
	g.Model = model
	g.oaiModel = m.oaiModel
	// XXX replace with a real tokenizer.
	// charsPerToken := 3.1
	// g.maxChunkLen = int(math.Floor(float64(m.TokenLimit) * charsPerToken))
	// XXX replace with a real tokenizer.
	// g.maxEmbeddingChunkLen = int(math.Floor(float64(8192) * charsPerToken))
	g.tokenLimit = m.TokenLimit
	// XXX 8192 hardcoded for the text-embedding-ada-002 model
	g.embeddingTokenLimit = 8192
	return
}

// SetModel sets the default chat completion model for queries.
func (g *Grokker) SetModel(model string) (oldModel string, err error) {
	defer Return(&err)
	model, _, err = g.models.findModel(model)
	Ck(err)
	oldModel, _, err = g.getModel()
	Ck(err)
	err = g.setup(model)
	Ck(err)
	return
}

// getModel returns the current model name and model_t from the db
func (g *Grokker) getModel() (model string, m *Model, err error) {
	defer Return(&err)
	model, m, err = g.models.findModel(g.Model)
	Ck(err)
	return
}

// Save saves the Grokker database as json data in an io.Writer.
func (g *Grokker) Save() (err error) {
	defer Return(&err)

	// open
	Debug("saving grok file")
	tmpfn := g.grokpath + ".tmp"
	fh, err := os.Create(tmpfn)

	// write
	data, err := json.Marshal(g)
	Ck(err)
	_, err = fh.Write(data)

	// close
	err = fh.Close()
	Ck(err)

	// move
	err = os.Rename(tmpfn, g.grokpath)
	Ck(err)
	Debug(" done!")

	return
}

// UpdateEmbeddings updates the embeddings for any documents that have
// changed since the last time the embeddings were updated.  It returns
// true if any embeddings were updated.
func (g *Grokker) UpdateEmbeddings() (update bool, err error) {
	defer Return(&err)
	// we use the timestamp of the grokfn as the last embedding update time.
	lastUpdate, err := g.Mtime()
	Ck(err)
	for _, doc := range g.Documents {
		// check if the document has changed.
		fi, err := os.Stat(g.AbsPath(doc))
		if os.IsNotExist(err) {
			// document has been removed; don't remove it from the
			// database, but don't update it either.  We don't want
			// to remove it from the database because it might be
			// on a different branch in e.g. git.
			// g.ForgetDocument(g.AbsPath(doc))
			// update = true
			err = nil
			continue
		}
		Ck(err)
		if fi.ModTime().After(lastUpdate) {
			// update the embeddings.
			Debug("updating embeddings for %s ...", doc.RelPath)
			updated, err := g.UpdateDocument(doc)
			Ck(err)
			Debug("done\n")
			update = update || updated
		}
	}
	// garbage collect any chunks that are no longer referenced.
	g.GC()
	return
}

// AddDocument adds a document to the Grokker database. It creates the
// embeddings for the document and adds them to the database.
func (g *Grokker) AddDocument(path string) (err error) {
	defer Return(&err)
	// assume we're in an arbitrary directory, so we need to
	// convert the path to an absolute path.
	absPath, err := filepath.Abs(path)
	Ck(err)
	// always convert path to a relative path for consistency
	relpath, err := filepath.Rel(g.Root, absPath)
	doc := &Document{
		RelPath: relpath,
	}
	// ensure the document exists
	_, err = os.Stat(g.AbsPath(doc))
	if os.IsNotExist(err) {
		err = fmt.Errorf("not found: %s", doc.RelPath)
		return
	}
	Ck(err)
	// find out if the document is already in the database.
	found := false
	for _, d := range g.Documents {
		if d.RelPath == doc.RelPath {
			found = true
			break
		}
	}
	if !found {
		// add the document to the database.
		g.Documents = append(g.Documents, doc)
	}
	// update the embeddings for the document.
	_, err = g.UpdateDocument(doc)
	Ck(err)
	return
}

// ForgetDocument removes a document from the Grokker database.
func (g *Grokker) ForgetDocument(path string) (err error) {
	defer Return(&err)
	// remove the document from the database.
	for i, d := range g.Documents {
		match := false
		// try comparing the paths directly first.
		if d.RelPath == path {
			match = true
		}
		// if that doesn't work, try comparing the absolute paths.
		relpath, err := filepath.Abs(path)
		Ck(err)
		if g.AbsPath(d) == relpath {
			match = true
		}
		if match {
			Debug("forgetting document %s ...", path)
			g.Documents = append(g.Documents[:i], g.Documents[i+1:]...)
			break
		}
	}
	// the document chunks are still in the database, but they will be
	// removed during garbage collection.
	return
}

// GC removes any chunks that are marked as stale
func (g *Grokker) GC() (err error) {
	defer Return(&err)
	// for each chunk, check if it is referenced by any document.
	// if not, remove it from the database.
	oldLen := len(g.Chunks)
	var keepChunks []*Chunk
	for _, chunk := range g.Chunks {
		if !chunk.stale {
			keepChunks = append(keepChunks, chunk)
		}
	}
	// replace the old chunks with the new chunks.
	g.Chunks = keepChunks
	newLen := len(g.Chunks)
	Debug("garbage collected %d chunks from the database", oldLen-newLen)
	return
}

// UpdateDocument updates the embeddings for a document and returns
// true if the document was updated.
func (g *Grokker) UpdateDocument(doc *Document) (updated bool, err error) {
	defer Return(&err)
	// XXX much of this code is inefficient and will be replaced
	// when we have a kv store.
	Debug("updating embeddings for %s ...", doc.RelPath)

	// mark all existing chunks as stale
	for _, chunk := range g.Chunks {
		if chunk.Document.RelPath == doc.RelPath {
			chunk.stale = true
		}
	}

	// break the current doc up into chunks.
	chunks, err := g.chunksFromDoc(doc)
	Ck(err)
	// For each chunk, ensure it exists in the database with the right
	// hash, offset, and length.  We'll get embeddings later.
	var newChunks []*Chunk
	for _, chunk := range chunks {
		newChunk := g.SetChunk(chunk)
		if newChunk != nil {
			updated = true
			newChunks = append(newChunks, newChunk)
		}
	}
	Debug("found %d new chunks", len(newChunks))
	// orphaned chunks will be garbage collected.

	// For each new chunk, generate an embedding using the
	// openai.Embedding.create() function. Store the embeddings for each
	// chunk in a data structure such as a list or dictionary.
	var newChunkStrings []string
	for _, chunk := range newChunks {
		Assert(chunk.Document.RelPath == doc.RelPath, "chunk document does not match")
		Assert(len(chunk.text) > 0, "chunk text is empty")
		Assert(chunk.Embedding == nil, "chunk embedding is not nil")
		Assert(chunk.stale == false, "chunk is stale")
		Assert(chunk.Hash != "", "chunk hash is empty")
		text, err := g.ChunkText(chunk, true)
		Ck(err)
		newChunkStrings = append(newChunkStrings, text)
	}
	embeddings, err := g.CreateEmbeddings(newChunkStrings)
	Ck(err)
	for i, chunk := range newChunks {
		chunk.Embedding = embeddings[i]
	}
	return
}

// SetChunk ensures that a chunk exists in the database with the right
// doc, hash, offset, and length, and unsets the stale bit.  It
// returns the chunk if it was added to the database, or nil if it was
// already in the database. The caller needs to set the embedding if
// newChunk is not nil.
func (g *Grokker) SetChunk(chunk *Chunk) (newChunk *Chunk) {
	// check if the chunk is already in the database.
	var foundChunk *Chunk
	for _, c := range g.Chunks {
		if c.Hash == chunk.Hash && c.Document.RelPath == chunk.Document.RelPath {
			foundChunk = c
			foundChunk.Offset = chunk.Offset
			foundChunk.Length = chunk.Length
			foundChunk.stale = false
		}
	}
	if foundChunk == nil {
		// add the chunk to the database.
		g.Chunks = append(g.Chunks, chunk)
		newChunk = chunk
		newChunk.stale = false
	}
	return
}

// Embeddings returns the embeddings for a slice of text chunks.
func (g *Grokker) CreateEmbeddings(texts []string) (embeddings [][]float64, err error) {
	// use github.com/fabiustech/openai library
	c := g.embeddingClient
	// simply return an empty list if there are no texts.
	if len(texts) == 0 {
		return
	}
	// iterate over the text chunks and create one or more embedding queries
	for i := 0; i < len(texts); {
		// add texts to the current query until we reach the token limit
		// i is the index of the first text in the current query
		// j is the index of the last text in the current query
		// XXX this is ugly, fragile, and needs to be tested and refactored
		totalLen := 0
		j := i
		for {
			var tokens []string
			_, tokens, err = g.tokenizer.Encode(texts[j])
			Ck(err)
			nextLen := len(tokens)
			// Debug("i=%d j=%d nextLen=%d totalLen=%d", i, j, nextLen, totalLen)
			Assert(nextLen > 0, texts)
			Assert(nextLen <= g.embeddingTokenLimit, "nextLen=%d embeddingTokenLimit=%d", nextLen, g.embeddingTokenLimit)
			if totalLen+nextLen >= g.embeddingTokenLimit {
				j--
				Debug("breaking because totalLen=%d nextLen=%d", totalLen, nextLen)
				break
			}
			totalLen += nextLen
			if j == len(texts)-1 {
				Debug("breaking because j=%d len(texts)=%d", j, len(texts))
				break
			}
			j++
		}
		Debug("i=%d j=%d totalLen=%d", i, j, totalLen)
		Assert(j >= i, "j=%d i=%d", j, i)
		Assert(totalLen > 0, "totalLen=%d", totalLen)
		inputs := texts[i : j+1]
		// double-check that the total length is within the limit and that
		// no individual text is too long.
		/*
			totalLen = 0
			for _, text := range inputs {
				totalLen += len(text)
				// Debug("len(text)=%d, totalLen=%d", len(text), totalLen)
				Assert(len(text) <= g.maxEmbeddingChunkLen, "text too long: %d", len(text))
			}
			Assert(totalLen <= g.maxEmbeddingChunkLen, "totalLen=%d maxEmbeddingChunkLen=%d", totalLen, g.maxEmbeddingChunkLen)
		*/
		req := &openai.EmbeddingRequest{
			Input: inputs,
			Model: fabius_models.AdaEmbeddingV2,
		}
		res, err := c.CreateEmbeddings(context.Background(), req)
		Ck(err)
		for _, em := range res.Data {
			embeddings = append(embeddings, em.Embedding)
		}
		i = j + 1
	}
	Debug("created %d embeddings", len(embeddings))
	Assert(len(embeddings) == len(texts))
	return
}

// chunksFromDoc returns a slice containing the chunks for a document.
func (g *Grokker) chunksFromDoc(doc *Document) (chunks []*Chunk, err error) {
	defer Return(&err)
	// read the document.
	buf, err := ioutil.ReadFile(g.AbsPath(doc))
	Ck(err)
	// break the document up into chunks.
	chunks = g.chunksFromString(doc, string(buf), g.embeddingTokenLimit)
	// add the document to each chunk.
	for _, chunk := range chunks {
		chunk.Document = doc
	}
	return
}

// splitIntoChunks splits a string into a slice of chunks using the
// given delimiter, returning each string as a partially populated
// Chunk with the offset set to the start of the string.
func splitIntoChunks(doc *Document, txt, delimiter string) (chunks []*Chunk) {
	start := 0
	i := 0
	for done := false; !done; i++ {
		var t string
		if i >= len(txt)-len(delimiter) {
			// we're at the end of the document: put the
			// remaining text into a chunk.
			t = txt[start:]
			done = true
		} else if txt[i:i+len(delimiter)] == delimiter {
			// found a delimiter
			t = txt[start : i+len(delimiter)]
		}
		// if we found a delimiter, then create a chunk
		// and add it to the list of chunks.
		if t != "" {
			chunk := NewChunk(doc, start, len(t), t)
			chunks = append(chunks, chunk)
			// start the next chunk after the delimiter
			start = i + len(delimiter)
			i = start
			continue
		}
	}
	return
}

// Tokens returns the tokens for a text segment.
func (g *Grokker) Tokens(text string) (tokens []string, err error) {
	defer Return(&err)
	_, tokens, err = g.tokenizer.Encode(text)
	Ck(err)
	return
}

// splitChunk recursively splits a Chunk into smaller chunks until
// each chunk is no longer than the token limit.
func (chunk *Chunk) splitChunk(g *Grokker, tokenLimit int) (newChunks []*Chunk) {
	// quick check: if the chunk is already short enough based on
	// the number of bytes, then we don't need to tokenize it.
	Assert(chunk.Length == len(chunk.text), "chunk length does not match text length")
	if chunk.Length < tokenLimit {
		newChunks = append(newChunks, chunk)
		return
	}
	// tokenize the chunk content
	tokens, err := g.Tokens(chunk.text)
	Ck(err)
	// if the chunk is short enough, then we're done
	if len(tokens) < tokenLimit {
		newChunks = append(newChunks, chunk)
		return
	}
	// split chunk into smaller segments of roughly equal size
	// XXX could be made smarter by splitting on sentence or context boundaries
	numChunks := int(math.Ceil(float64(len(tokens)) / float64(tokenLimit)))
	chunkSize := int(math.Ceil(float64(chunk.Length) / float64(numChunks)))
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > chunk.Length {
			end = chunk.Length
		}
		docOffset := chunk.Offset + start
		subChunk := NewChunk(chunk.Document, docOffset, end-start, chunk.text[start:end])
		// recurse
		newSubChunks := subChunk.splitChunk(g, tokenLimit)
		newChunks = append(newChunks, newSubChunks...)
	}
	return
}

// chunksFromString splits a string into a slice of Chunks.  If doc is
// not nil, it is used to set the Document field of each chunk.  Each
// chunk will be no longer than tokenLimit tokens.
func (g *Grokker) chunksFromString(doc *Document, txt string, tokenLimit int) (chunks []*Chunk) {
	Assert(tokenLimit > 0)

	// split the text into paragraphs
	// XXX splitting on paragraphs is not ideal.  smarter splitting
	// might look at the structure of the text and split on
	// sections, chapters, etc.  it might also be useful to include
	// metadata such as file names.
	paragraphs := splitIntoChunks(doc, txt, "\n\n")

	for _, chunk := range paragraphs {
		// ensure no paragraph is longer than the token limit
		subChunks := chunk.splitChunk(g, tokenLimit)
		chunks = append(chunks, subChunks...)
	}
	return
}

// (4) When a question is asked, use the embeddings of the document
// chunks to find the most relevant passages for the question. You can
// use a similarity measure such as cosine similarity to compare the
// embeddings of the question and each document chunk, and return the
// chunks with the highest similarity scores.

// FindChunks returns the most relevant chunks for a query, limited by tokenLimit.
func (g *Grokker) FindChunks(query string, tokenLimit int) (chunks []*Chunk, err error) {
	defer Return(&err)
	// get the embeddings for the query.
	embeddings, err := g.CreateEmbeddings([]string{query})
	Ck(err)
	queryEmbedding := embeddings[0]
	// find the most similar chunks.
	chunks = g.SimilarChunks(queryEmbedding, tokenLimit)
	return
}

// SimilarChunks returns the most similar chunks to an embedding,
// limited by tokenLimit.
func (g *Grokker) SimilarChunks(embedding []float64, tokenLimit int) (chunks []*Chunk) {
	Debug("chunks in database: %d", len(g.Chunks))
	// find the most similar chunks.
	type Sim struct {
		chunk *Chunk
		score float64
	}
	sims := make([]Sim, 0, len(g.Chunks))
	for _, chunk := range g.Chunks {
		score := Similarity(embedding, chunk.Embedding)
		sims = append(sims, Sim{chunk, score})
	}
	// sort the chunks by similarity.
	sort.Slice(sims, func(i, j int) bool {
		return sims[i].score > sims[j].score
	})
	// return the top chunks until we reach the token limit.
	var totalTokens int
	for _, sim := range sims {
		// tokenize the text XXX do this before storing in the
		// database instead of on every retrieval. will need to ensure
		// that the chunk header is handled correctly.
		text, err := g.ChunkText(sim.chunk, true)
		Ck(err)
		tokens, err := g.Tokens(text)
		Ck(err)
		totalTokens += len(tokens)
		if totalTokens > tokenLimit {
			break
		}
		chunks = append(chunks, sim.chunk)
	}
	Debug("found %d similar chunks", len(chunks))
	return
}

// Similarity returns the cosine similarity between two embeddings.
func Similarity(a, b []float64) float64 {
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

// (5) Provide the most relevant document chunks to the
// openai.Completion.create() function as additional context for
// generating a response. This will allow the model to better
// understand the context of the question and generate a more relevant
// response.

// Answer returns the answer to a question.
func (g *Grokker) Answer(question string, global bool) (resp string, err error) {
	defer Return(&err)
	// tokenize the question
	qtokens, err := g.Tokens(question)
	Ck(err)
	maxTokens := int(float64(g.tokenLimit)*0.5) - len(qtokens)
	context, err := g.getContext(question, maxTokens)
	Ck(err)
	// generate the answer.
	respmsg, err := g.Generate(sysMsgChat, question, context, global)
	resp = respmsg.Choices[0].Message.Content
	return
}

// Revise returns revised text based on input text.
func (g *Grokker) Revise(in string, global, sysmsgin bool) (out, sysmsg string, err error) {
	defer Return(&err)

	// tokenize the entire input
	inTokens, err := g.Tokens(in)

	var body string
	if sysmsgin {
		// split input into sysmsg and txt
		paragraphs := strings.Split(in, "\n\n")
		if len(paragraphs) < 2 {
			err = fmt.Errorf("input must contain at least two paragraphs")
			return
		}
		sysmsg = paragraphs[0]
		// in = strings.Join(paragraphs[1:], "\n\n")
		body = strings.Join(paragraphs, "\n\n")
	} else {
		sysmsg = sysMsgRevise
	}

	// get context
	maxTokens := int(float64(g.tokenLimit)*0.5) - len(inTokens)
	context, err := g.getContext(body, maxTokens)
	Ck(err)

	// generate the answer.
	resp, err := g.Generate(sysmsg, in, context, global)
	Ck(err)
	if sysmsgin {
		out = Spf("%s\n\n%s", sysmsg, resp.Choices[0].Message.Content)
	} else {
		out = resp.Choices[0].Message.Content
	}
	return
}

// getContext returns the context for a query.
func (g *Grokker) getContext(query string, tokenLimit int) (context string, err error) {
	defer Return(&err)
	// get chunks, sorted by similarity to the query.
	chunks, err := g.FindChunks(query, tokenLimit)
	Ck(err)
	for _, chunk := range chunks {
		text, err := g.ChunkText(chunk, true)
		Ck(err)
		context += text
	}
	Debug("using %d chunks as context", len(chunks))
	return
}

// Continue returns a continuation of the input text.
func (g *Grokker) Continue(in string, global bool) (out, sysmsg string, err error) {
	defer Return(&err)
	sysmsg = sysMsgContinue
	// tokenize sysmsg
	_, sysmsgTokens, err := g.tokenizer.Encode(sysmsg)
	Ck(err)
	// tokenize input
	_, inTokens, err := g.tokenizer.Encode(in)
	Ck(err)
	// get chunks, sorted by similarity to the txt.
	tokenLimit := int(float64(g.tokenLimit)*0.4) - len(sysmsgTokens) - len(inTokens)
	context, err := g.getContext(in, tokenLimit)
	Ck(err)
	// generate the answer.
	resp, err := g.Generate(sysmsg, in, context, global)
	Ck(err)
	out = resp.Choices[0].Message.Content
	return
}

// Use the openai.Completion.create() function to generate a
// response to the question. You can use the prompt parameter to
// provide the question, and the max_tokens parameter to limit the
// length of the response.

// var promptTmpl = `You are a helpful assistant.  Answer the following question and summarize the context:
// var promptTmpl = `You are a helpful assistant.
var promptTmpl = `{{.Question}}

Context:
{{.Context}}`

var sysMsgChat = "You are an expert knowledgable in the provided context.  I will provide you with context, then you will respond with an acknowledgement, then I will ask you a question about the context, then you will provide me with an answer."

var sysMsgRevise = "You are an expert knowledgable in the provided context.  I will provide you with context, then you will respond with an acknowledgement, then I will provide you with a block of text.  You will revise the block of text based on the information in the context, maintaining the same style, vocabulary, and reading level."

var sysMsgContinue = "You are an expert knowledgable in the provided context.  I will provide you with context, then you will respond with an acknowledgement, then I will provide you with a block of text.  You will continue the block of text based on the information in the context, maintaining the same style, vocabulary, and reading level."

// Generate returns the answer to a question.
func (g *Grokker) Generate(sysmsg, question, ctxt string, global bool) (resp oai.ChatCompletionResponse, err error) {
	defer Return(&err)

	// XXX don't exceed max tokens
	messages := []oai.ChatCompletionMessage{
		{
			Role:    oai.ChatMessageRoleSystem,
			Content: sysmsg,
		},
	}

	// first get global knowledge
	if global {
		messages = append(messages, oai.ChatCompletionMessage{
			Role:    oai.ChatMessageRoleUser,
			Content: question,
		})
		resp, err = g.Chat(messages)
		Ck(err)
		// add the response to the messages.
		messages = append(messages, oai.ChatCompletionMessage{
			Role:    oai.ChatMessageRoleAssistant,
			Content: resp.Choices[0].Message.Content,
		})
	}

	// add context from local sources
	if len(ctxt) > 0 {
		messages = append(messages, []oai.ChatCompletionMessage{
			{
				Role:    oai.ChatMessageRoleUser,
				Content: Spf("Context:\n\n%s", ctxt),
			},
			{
				Role:    oai.ChatMessageRoleAssistant,
				Content: "Great! I've read the context.",
			},
		}...)
	}

	// now ask the question
	messages = append(messages, oai.ChatCompletionMessage{
		Role:    oai.ChatMessageRoleUser,
		Content: question,
	})

	// get the answer
	resp, err = g.Chat(messages)
	Ck(err, "context length: %d", len(ctxt))

	// fmt.Println(resp.Choices[0].Message.Content)
	// Pprint(messages)
	// Pprint(resp)
	return
}

// Chat uses the openai API to continue a conversation given a
// (possibly synthesized) message history.
func (g *Grokker) Chat(messages []oai.ChatCompletionMessage) (resp oai.ChatCompletionResponse, err error) {
	defer Return(&err)

	model := g.oaiModel
	Debug("chat model: %s", model)
	Debug("chat: messages: %v", messages)

	// use 	"github.com/sashabaranov/go-openai"
	client := g.chatClient
	resp, err = client.CreateChatCompletion(
		context.Background(),
		oai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	Ck(err, "%#v", messages)
	totalBytes := 0
	for _, msg := range messages {
		totalBytes += len(msg.Content)
	}
	totalBytes += len(resp.Choices[0].Message.Content)
	ratio := float64(totalBytes) / float64(resp.Usage.TotalTokens)
	Debug("chat response: %s", resp)
	Debug("total tokens: %d  char/token ratio: %.1f\n", resp.Usage.TotalTokens, ratio)
	return
}

// ListDocuments returns a list of all documents in the knowledge base.
// XXX this is a bit of a hack, since we're using the document name as
// the document ID.
// XXX this is also a bit of a hack since we're trying to make this
// work for multiple versions -- we should be able to simplify this
// after migration is automatic during Load().
func (g *Grokker) ListDocuments() (paths []string) {
	for _, doc := range g.Documents {
		path := doc.Path
		v100, err := semver.Parse([]byte("1.0.0"))
		current, err := semver.Parse([]byte(g.Version))
		Ck(err)
		if semver.Cmp(current, v100) > 0 {
			path = doc.RelPath
		}
		paths = append(paths, path)
	}
	return
}

// ListModels lists the available models.
func (g *Grokker) ListModels() (models []*Model, err error) {
	defer Return(&err)
	for _, model := range g.models.Available {
		models = append(models, model)
	}
	return
}

// RefreshEmbeddings refreshes the embeddings for all documents in the
// database.
func (g *Grokker) RefreshEmbeddings() (err error) {
	defer Return(&err)
	// regenerate the embeddings for each document.
	for _, doc := range g.Documents {
		Fpf(os.Stderr, "refreshing embeddings for %s\n", doc.RelPath)
		// remove file from list if it doesn't exist.
		absPath := g.AbsPath(doc)
		Debug("absPath: %s", absPath)
		_, err := os.Stat(absPath)
		Debug("stat err: %v", err)
		if os.IsNotExist(err) {
			// remove the document from the database.
			g.ForgetDocument(doc.RelPath)
			continue
		}
		_, err = g.UpdateDocument(doc)
		Ck(err)
	}
	g.GC()
	return
}

/*
var GitSummaryPrompt = `
Summarize the context into a single line of 60 characters or less.  Add nothing else.  Use present tense.  Do not quote.
`
var GitCommitPrompt = `
Summarize the bullet points found in the context into a single line of 60 characters or less.  Add nothing else.  Use present tense. Do not quote.
`
*/

var GitSummaryPrompt = `
Describe the most important item in the context in a single line of 60 characters or less.  Add nothing else.  Use present tense.  Do not quote.  Use present tense.  Use active voice.  Do not use passive voice.
`

var GitCommitPrompt = `
Describe the most important bullet point in the context in a single line of 60 characters or less.  Add nothing else.  Use present tense.  Do not quote.  Use present tense.  Use active voice.  Do not use passive voice.
`

var GitDiffPrompt = `
Summarize the bullet points and 'git diff' fragments found in the context into bullet points to be used in the body of a git commit message.  Add nothing else. Use present tense.  Use active voice.  Do not use passive voice.
`

// GitCommitMessage generates a git commit message given a diff. It
// appends a reasonable prompt, and then uses the result as a grokker
// query.
func (g *Grokker) GitCommitMessage(diff string) (msg string, err error) {
	defer Return(&err)

	// summarize the diff
	sumLines, msg, err := g.summarizeDiff(diff)
	Ck(err)

	// summarize the sumLines to create the first line of the commit
	// message
	resp, err := g.Generate(sysMsgChat, GitSummaryPrompt, sumLines, false)
	Ck(err)
	summary := resp.Choices[0].Message.Content

	// glue it all together
	msg = Spf("%s\n\n%s", summary, msg)

	return
}

// summarizeDiff recursively summarizes a diff until the summary is
// short enough to be used as a prompt.
func (g *Grokker) summarizeDiff(diff string) (sumlines string, diffSummary string, err error) {
	defer Return(&err)
	maxTokens := int(float64(g.tokenLimit) * .7)
	// split the diff on filenames
	fileChunks := strings.Split(diff, "diff --git")
	// split each file chunk into smaller chunks
	for _, fileChunk := range fileChunks {
		// skip empty chunks
		if len(fileChunk) == 0 {
			continue
		}
		// get the filenames (they were right after the "diff --git"
		// string, on the same line)
		lines := strings.Split(fileChunk, "\n")
		var fns string
		if len(lines) > 0 {
			fns = lines[0]
		} else {
			fns = "a b"
		}
		var fileSummary string
		if len(fns) > 0 {
			fileSummary = Spf("summary of diff --git %s\n", fns)
		}
		chunks := g.chunksFromString(nil, fileChunk, maxTokens)
		// summarize each chunk
		for _, chunk := range chunks {
			// format the chunk
			context := Spf("diff --git %s\n%s", fns, chunk.text)
			resp, err := g.Generate(sysMsgChat, GitDiffPrompt, context, false)
			Ck(err)
			fileSummary = Spf("%s\n%s", fileSummary, resp.Choices[0].Message.Content)
		}
		// XXX recurse here to glue the summaries together for a given
		// file?

		// get a summary line of the changes for this file
		resp, err := g.Generate(sysMsgChat, GitCommitPrompt, fileSummary, false)
		Ck(err)
		sumLine := resp.Choices[0].Message.Content
		// append the summary line to the list of summary lines
		sumlines = Spf("%s\n%s", sumlines, sumLine)
		// append sumLine and the diff for this file to the summary
		// of the changes for all files
		diffSummary = Spf("%s\n\n%s\n\n%s", diffSummary, sumLine, fileSummary)
	}
	/*
		// if the summary is too long, recurse
		if len(diffSummary) > int(maxLen) {
			// recurse
			Fpf(os.Stderr, "diff summary too long (%d bytes), recursing\n", len(diffSummary))
			diffSummary, err = g.summarizeDiff(diffSummary)
		}
	*/
	return
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) (err error) {
	defer Return(&err)
	// open src file
	srcfh, err := os.Open(src)
	Ck(err)
	defer srcfh.Close()
	// ensure dst file does not exist
	_, err = os.Stat(dst)
	if err == nil {
		Fpf(os.Stderr, "Error: %s already exists\n", dst)
		os.Exit(1)
	}
	// open dst file with same mode as src
	fi, err := srcfh.Stat()
	Ck(err)
	dstfh, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
	Ck(err)
	defer dstfh.Close()
	// copy
	_, err = io.Copy(dstfh, srcfh)
	Ck(err)
	return
}
