package ytaccount

import "context"

// ProbeBrowse is a temporary recon hook; it goes away once Home lands.
func (c *Client) ProbeBrowse(ctx context.Context, id string) (map[string]any, error) {
	return c.call(ctx, "browse", map[string]any{"browseId": id})
}
