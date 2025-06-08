package probe

import "context"

type Probe interface {
	Run(ctx context.Context, target string)
}
