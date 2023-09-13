package sdk

import (
	"errors"
	"strconv"
	"strings"
)

//
// OCIImageReference
//

type OCIImageReference struct {
	Artifact string `ard:"artifact"`

	Reference string `ard:"reference"`

	Host  string `ard:"host"`
	Image string `ard:"image"`
	Tag   string `ard:"tag"`

	Port            int64  `ard:"port"`
	Repository      string `ard:"repository"`
	DigestAlgorithm string `ard:"digest-algorithm"`
	DigestHex       string `ard:"digest-hex"`
}

func (self OCIImageReference) Validate() error {
	if self.Artifact != "" {
		if (self.Reference != "") || (self.Host != "") || (self.Image != "") || (self.Tag != "") || (self.Port != 0) || (self.Repository != "") || (self.DigestAlgorithm != "") || (self.DigestHex != "") {
			return errors.New("invalid OCI image reference: `artifact` incompatible with other properties")
		}
	}

	if self.Reference != "" {
		if (self.Artifact != "") || (self.Host != "") || (self.Image != "") || (self.Tag != "") || (self.Port != 0) || (self.Repository != "") || (self.DigestAlgorithm != "") || (self.DigestHex != "") {
			return errors.New("invalid OCI image reference: `reference` incompatible with other properties")
		}
	}

	if self.Image != "" {
		if (self.Host == "") || (self.Tag == "") {
			return errors.New("invalid OCI image reference: `image`, `host`, and `tag` must be set together")
		}

		if (self.Artifact != "") || (self.Reference != "") {
			return errors.New("invalid OCI image reference: `image` incompatible with `artifact` and `reference`")
		}
	}
	if self.Host != "" {
		if (self.Image == "") || (self.Tag == "") {
			return errors.New("invalid OCI image reference: `image`, `host`, and `tag` must be set together")
		}

		if (self.Artifact != "") || (self.Reference != "") {
			return errors.New("invalid OCI image reference: `image` incompatible with `artifact` and `reference`")
		}
	}
	if self.Tag != "" {
		if (self.Host == "") || (self.Image == "") {
			return errors.New("invalid OCI image reference: `image`, `host`, and `tag` must be set together")
		}

		if (self.Artifact != "") || (self.Reference != "") {
			return errors.New("invalid OCI image reference: `image` incompatible with `artifact` and `reference`")
		}
	}

	if (self.DigestAlgorithm != "") && (self.DigestHex == "") {
		return errors.New("invalid OCI image reference: `digest-algorithm` and `digest-hex` must be set together")
	}
	if (self.DigestHex != "") && (self.DigestAlgorithm == "") {
		return errors.New("invalid OCI image reference: `digest-algorithm` and `digest-hex` must be set together")
	}

	return nil
}

// ([fmt.Stringer] interface)
func (self OCIImageReference) String() string {
	// [host[:port]/][repository/]image[:tag][@digest-algorithm:digest-hex]

	if self.Reference != "" {
		return self.Reference
	}

	var s strings.Builder

	if self.Host != "" {
		s.WriteString(self.Host)
		if self.Port != 0 {
			s.WriteRune(':')
			s.WriteString(strconv.Itoa(int(self.Port)))
		}
		s.WriteRune('/')
	}

	if self.Repository != "" {
		s.WriteString(self.Repository)
		s.WriteRune('/')
	}

	s.WriteString(self.Image)

	if self.Tag != "" {
		s.WriteRune(':')
		s.WriteString(self.Tag)
	}

	if (self.DigestAlgorithm != "") && (self.DigestHex != "") {
		s.WriteRune('@')
		s.WriteString(self.DigestAlgorithm)
		s.WriteRune(':')
		s.WriteString(self.DigestHex)
	}

	return s.String()
}
