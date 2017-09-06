// Policy evaluation for prSignedBaseLayer.

package signature

import (
	"github.com/Sirupsen/logrus"
	"github.com/containers/image/types"
)

func (pr *prSignedBaseLayer) isSignatureAuthorAccepted(image types.UnparsedImage, sig []byte) (signatureAcceptanceResult, *Signature, error) {
	return sarUnknown, nil, nil
}

func (pr *prSignedBaseLayer) isRunningImageAllowed(image types.UnparsedImage) (bool, error) {
	// FIXME? Reject this at policy parsing time already?
	logrus.Errorf("signedBaseLayer not implemented yet!")
	return false, PolicyRequirementError("signedBaseLayer not implemented yet!")
}
