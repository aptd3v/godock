package commitoptions

import (
	"github.com/docker/docker/api/types/container"
)

type CommitOptionsFn func(options *container.CommitOptions)

func Reference(reference string) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		options.Reference = reference
	}
}
func Comment(comment string) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		options.Comment = comment
	}
}
func Author(author string) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		options.Author = author
	}
}
func Changes(changes ...string) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		if options.Changes == nil {
			options.Changes = []string{}
		}
		options.Changes = append(options.Changes, changes...)
	}
}
func Pause(pause bool) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		options.Pause = pause
	}
}
func Config(config *container.Config) CommitOptionsFn {
	return func(options *container.CommitOptions) {
		options.Config = config
	}
}
