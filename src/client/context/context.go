package context

import (
	"scaffold/client/auth"
)

func DoContext(profile, context string) {
	p := auth.ReadProfile(profile)

	p.Cascade = context

	auth.WriteProfile(profile, p)
}
