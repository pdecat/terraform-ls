package handlers

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	lsp "github.com/sourcegraph/go-lsp"
)

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    lsp.Range     `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func (h *logHandler) TextDocumentHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*Hover, error) {
	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return nil, err
	}

	cc, err := lsctx.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	df, err := lsctx.DecoderFinder(ctx)
	if err != nil {
		return nil, err
	}

	file, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}

	d, err := df.DecoderForDir(file.Dir())
	if err != nil {
		return nil, fmt.Errorf("finding compatible decoder failed: %w", err)
	}

	fPos, err := ilsp.FilePositionFromDocumentPosition(params, file)
	if err != nil {
		return nil, err
	}

	h.logger.Printf("Looking for hover data at %q -> %#v", file.Filename(), fPos.Position())
	data, err := d.HoverAtPos(file.Filename(), fPos.Position())
	if err != nil {
		return nil, err
	}

	supportsMarkdown := false
	contentFmt := cc.TextDocument.Hover.ContentFormat
	if len(contentFmt) > 0 {
		if contentFmt[0] == "markdown" {
			supportsMarkdown = true
		}
	}

	return convertHoverData(data, supportsMarkdown), nil
}

func convertHoverData(data *lang.HoverData, markdown bool) *Hover {
	// TODO: markdown stripping
	content := MarkupContent{
		Kind:  "markdown",
		Value: data.Content.Value,
	}
	return &Hover{
		Contents: content,
		Range:    ilsp.HCLRangeToLSP(data.Range),
	}
}
