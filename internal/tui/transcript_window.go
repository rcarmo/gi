package tui

import (
	"container/list"
	"encoding/json"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	gotui "github.com/grindlemire/go-tui"
)

// Transcript windowing (#34): a frame lays out only the transcript blocks
// that intersect the viewport. The rest are fixed-height spacers sized from
// cached block heights, so the scroll container's content height, scroll
// offsets and "stick to bottom" behave exactly as with every block laid
// out, while a frame costs what the screen shows instead of the whole
// session.

// transcriptWindowMargin is how many extra rows above and below the
// viewport are laid out, so small scrolls stay smooth.
const transcriptWindowMargin = 8

// blockHeightKey identifies a block's rendered height: its content, the
// spacing context (previous kind), width and theme.
func blockHeightKey(block transcriptRenderableBlock, previousKind string, width int) uint64 {
	h := fnv.New64a()
	write := func(s string) {
		_, _ = h.Write([]byte(s))
		_, _ = h.Write([]byte{0})
	}
	write(block.Kind)
	write(block.Key)
	write(block.Header)
	write(block.Subheader)
	for _, line := range block.Body {
		write(line)
	}
	write(block.MarkdownSource)
	write(block.Footer)
	write(block.Status)
	write(block.ToolPath)
	write(block.ToolArg)
	write(block.StartedAt)
	write(block.EndedAt)
	if block.DurationMS != nil {
		write(strconv.FormatInt(*block.DurationMS, 10))
	}
	if block.ToolContent != nil {
		write(*block.ToolContent)
	}
	write(block.ToolRange)
	write(block.ToolNotice)
	if block.EditDiff != nil {
		write("diff:" + *block.EditDiff)
	}
	write(block.EditError)
	if len(block.Calls) > 0 {
		raw, _ := json.Marshal(block.Calls)
		write(string(raw))
	}
	write(block.FullOutputPath)
	write(strconv.FormatBool(block.Expanded) + strconv.FormatBool(block.Expandable) + strconv.FormatBool(block.Selected) +
		strconv.FormatBool(block.Static) + strconv.FormatBool(block.PreviewTail) + strconv.Itoa(block.PreviewLimit))
	write(previousKind)
	write(strconv.Itoa(width))
	write(piActiveTheme + "#" + strconv.Itoa(piThemeGeneration))
	return h.Sum64()
}

// cachedTranscriptBlock is a rendered block reused across frames: its
// element (with go-tui's wrap caches) and its click targets.
type cachedTranscriptBlock struct {
	el   *gotui.Element
	refs []transcriptBlockHitTarget
}

// Memory follows the screen, not the session (#31): every block's height is
// kept (a few bytes each, so offscreen blocks are measured once), but
// rendered elements only for the most recently shown blocks, which covers
// the viewport and its margins on any terminal.
const (
	transcriptHeightCacheMax  = 65536
	transcriptElementCacheMax = 256
)

// transcriptBlockCache holds block heights and an LRU of rendered blocks.
// A nil cache is valid and empty.
type transcriptBlockCache struct {
	heights  map[uint64]int
	elements map[uint64]*list.Element // of *elementCacheEntry
	order    *list.List               // most recently used first
}

type elementCacheEntry struct {
	key   uint64
	block *cachedTranscriptBlock
}

func (b *transcriptBlockCache) height(key uint64) (int, bool) {
	if b == nil {
		return 0, false
	}
	h, ok := b.heights[key]
	return h, ok
}

func (b *transcriptBlockCache) element(key uint64) (*cachedTranscriptBlock, bool) {
	if b == nil {
		return nil, false
	}
	e, ok := b.elements[key]
	if !ok {
		return nil, false
	}
	b.order.MoveToFront(e)
	return e.Value.(*elementCacheEntry).block, true
}

func (b *transcriptBlockCache) putHeight(key uint64, h int) {
	if len(b.heights) >= transcriptHeightCacheMax {
		b.heights = map[uint64]int{}
	}
	b.heights[key] = h
}

func (b *transcriptBlockCache) putElement(key uint64, block *cachedTranscriptBlock) {
	if e, ok := b.elements[key]; ok {
		e.Value.(*elementCacheEntry).block = block
		b.order.MoveToFront(e)
		return
	}
	b.elements[key] = b.order.PushFront(&elementCacheEntry{key, block})
	for b.order.Len() > transcriptElementCacheMax {
		oldest := b.order.Back()
		b.order.Remove(oldest)
		delete(b.elements, oldest.Value.(*elementCacheEntry).key)
	}
}

func (b *transcriptBlockCache) elementCount() int {
	if b == nil {
		return 0
	}
	return b.order.Len()
}

func (b *transcriptBlockCache) heightCount() int {
	if b == nil {
		return 0
	}
	return len(b.heights)
}

func (c *chatTUI) ensureBlockCache() *transcriptBlockCache {
	if c.blockCache == nil {
		c.blockCache = &transcriptBlockCache{heights: map[uint64]int{}, elements: map[uint64]*list.Element{}, order: list.New()}
	}
	return c.blockCache
}

// transcriptBlockHeight is a block's height, rendering it only when it was
// never measured; the rendered element goes to the LRU, not a permanent map.
func (c *chatTUI) transcriptBlockHeight(block *transcriptRenderableBlock, previousKind string, key uint64, width int) int {
	if block.Status != "running" && block.Kind != "thinking_indicator" {
		if h, ok := c.blockCache.height(key); ok {
			return h
		}
	}
	_, h := c.transcriptBlockElement(block, previousKind, key, width, false)
	return h
}

// transcriptBlockElement renders a block, or reuses its element from an
// earlier frame. Running blocks (spinner, elapsed time) are always
// rendered. register adds the block's click targets for this frame.
func (c *chatTUI) transcriptBlockElement(block *transcriptRenderableBlock, previousKind string, key uint64, width int, register bool) (*gotui.Element, int) {
	cacheable := block.Status != "running" && block.Kind != "thinking_indicator"
	if cacheable {
		if hit, ok := c.blockCache.element(key); ok {
			if register {
				c.transcriptBlockRefs = append(c.transcriptBlockRefs, hit.refs...)
			}
			h, _ := c.blockCache.height(key)
			return hit.el, h
		}
	}
	before := len(c.transcriptBlockRefs)
	el := c.renderTranscriptBlockAfter(*block, previousKind)
	refs := append([]transcriptBlockHitTarget(nil), c.transcriptBlockRefs[before:]...)
	if !register {
		c.transcriptBlockRefs = c.transcriptBlockRefs[:before]
	}
	h := el.HeightForWidth(width)
	if cacheable {
		cache := c.ensureBlockCache()
		cache.putHeight(key, h)
		cache.putElement(key, &cachedTranscriptBlock{el: el, refs: refs})
	}
	return el, h
}

// addTranscriptWindow adds the blocks that intersect the viewport to the
// transcript container, with spacers standing in for the rest.
func (c *chatTUI) addTranscriptWindow(transcript *gotui.Element, blocks []transcriptRenderableBlock, width, viewport int) {
	// Per-block placement; blocks are referenced by index, not copied.
	type placed struct {
		previous string
		top, h   int
	}
	keys := c.transcriptBlockKeys(blocks, width)
	items := make([]placed, len(blocks))
	total := 0
	previousKind := ""
	for i := range blocks {
		block := &blocks[i]
		h := c.transcriptBlockHeight(block, previousKind, keys[i], width)
		items[i] = placed{previous: previousKind, top: total, h: h}
		total += h
		if block.Kind != "thinking_indicator" {
			previousKind = block.Kind
		}
	}
	offset := c.transcriptScroll
	if c.stickToBottom {
		offset = total - viewport
	}
	offset = max(0, min(offset, max(0, total-viewport)))
	from, to := offset-transcriptWindowMargin, offset+viewport+transcriptWindowMargin
	spacer := func(rows int) {
		if rows > 0 {
			transcript.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(rows)))
		}
	}
	pending := 0 // rows of skipped blocks not yet emitted as a spacer
	for i, it := range items {
		if it.top+it.h <= from || it.top >= to {
			pending += it.h
			continue
		}
		spacer(pending)
		pending = 0
		el, _ := c.transcriptBlockElement(&blocks[i], it.previous, keys[i], width, true)
		transcript.AddChild(el)
	}
	spacer(pending)
}

// transcriptBlocksMemo caches the block list built from the transcript and
// the blocks' cache keys. Transcript lines are immutable strings, so
// comparing their data pointers detects any change without hashing text.
type transcriptBlocksMemo struct {
	lines      []string
	state      string
	blocks     []transcriptRenderableBlock
	generation uint64
	keysWidth  int
	keysGen    uint64
	keys       []uint64
}

func sameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) || unsafe.StringData(a[i]) != unsafe.StringData(b[i]) {
			return false
		}
	}
	return true
}

// transcriptBlockState is everything besides the lines that the block list
// depends on.
func (c *chatTUI) transcriptBlockState() string {
	var b strings.Builder
	b.WriteString(c.selectedTranscriptBlock)
	b.WriteByte(0)
	b.WriteString(c.cfg.AssistantName)
	b.WriteByte(0)
	b.WriteString(piActiveTheme + "#" + strconv.Itoa(piThemeGeneration))
	b.WriteByte(0)
	b.WriteString(strconv.Itoa(c.currentScrollbackLimit()))
	b.WriteByte(0)
	b.WriteString(strconv.Itoa(c.outputWidth))
	if !c.regularMode && c.startupHeaderShown() {
		b.WriteString("\x00h" + strings.Join(c.startupSignature(), "\x01"))
	}
	for _, k := range sortedKeys(c.transcriptExpanded) {
		b.WriteString("\x00e" + k + strconv.FormatBool(c.transcriptExpanded[k]))
	}
	for _, k := range sortedKeys(c.extensionToolModes) {
		b.WriteString("\x00m" + k + "=" + c.extensionToolModes[k])
	}
	return b.String()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// transcriptBlocks returns the visible transcript's blocks, rebuilt only
// when the transcript lines or the block state changed.
func (c *chatTUI) transcriptBlocks() []transcriptRenderableBlock {
	state := c.transcriptBlockState()
	m := &c.blocksMemo
	if m.blocks != nil && m.state == state && sameLines(m.lines, c.transcript) {
		return m.blocks
	}
	m.lines = append(m.lines[:0], c.transcript...)
	m.state = state
	m.blocks = c.buildTranscriptRenderableBlocks(c.visibleTranscript())
	m.generation++
	return m.blocks
}

// transcriptBlockKeys returns the cache keys of blocks at width, memoized
// for the current block list.
func (c *chatTUI) transcriptBlockKeys(blocks []transcriptRenderableBlock, width int) []uint64 {
	m := &c.blocksMemo
	memoized := len(blocks) > 0 && len(m.blocks) == len(blocks) && &m.blocks[0] == &blocks[0]
	if memoized && m.keysGen == m.generation && m.keysWidth == width && len(m.keys) == len(blocks) {
		return m.keys
	}
	keys := make([]uint64, len(blocks))
	previousKind := ""
	for i, block := range blocks {
		keys[i] = blockHeightKey(block, previousKind, width)
		if block.Kind != "thinking_indicator" {
			previousKind = block.Kind
		}
	}
	if memoized {
		m.keys, m.keysGen, m.keysWidth = keys, m.generation, width
	}
	return keys
}
