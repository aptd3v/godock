package listoptions

import (
	"github.com/aptd3v/godock/pkg/godock"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

/*
WithAll sets the all parameter for the list options.

	client.ContainerList(ctx, listoptions.WithAll(true))
*/
func WithAll(all bool) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.All = all
	}
}

/*
WithSize sets the size parameter for the list options.

	client.ContainerList(ctx, listoptions.WithSize(true))
*/
func WithSize(size bool) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.Size = size
	}
}

/*
WithLatest sets the latest parameter for the list options.

	client.ContainerList(ctx, listoptions.WithLatest(true))
*/
func WithLatest(latest bool) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.Latest = latest
	}
}

/*
WithSince sets the since parameter for the list options.

	client.ContainerList(ctx, listoptions.WithSince("2021-01-01"))
*/
func WithSince(since string) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.Since = since
	}
}

/*
WithBefore sets the before parameter for the list options.

	client.ContainerList(ctx, listoptions.WithBefore("2021-01-01"))
*/
func WithBefore(before string) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.Before = before
	}
}

/*
WithLimit sets the limit for the list options.

	client.ContainerList(ctx, listoptions.WithLimit(10))
*/
func WithLimit(limit int) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		opts.Limit = limit
	}
}

/*
WithFilters adds a filter to the list options.
If the filter already exists, it will be overwritten.

	client.ContainerList(ctx,
	listoptions.WithFilters("status", "running"),
	listoptions.WithFilters("label", "test"),
	)
*/
func WithFilters(key, value string) godock.ListOptionFn {
	return func(opts *container.ListOptions) {
		if opts.Filters.Len() == 0 {
			opts.Filters = filters.NewArgs()
		}
		opts.Filters.Add(key, value)
	}
}
