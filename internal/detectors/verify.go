package detectors

import (
	"net/url"
	"strings"

	"encoding/base64"
	"encoding/hex"

	"github.com/redactyl/redactyl/internal/types"
	v "github.com/redactyl/redactyl/internal/validate"
)

// verifyFilters defines optional soft-verify checks per detector.
// In "safe" mode, these are applied to further tighten acceptance.
// They must be purely local (no network calls, no exfiltration).
var verifyFilters = map[string]func(f types.Finding) (types.Finding, bool){
	// Stripe: require slightly longer tail in safe mode (>= 28)
	"stripe_secret": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "sk_live_") && len(m) >= len("sk_live_")+28 {
			return f, true
		}
		return f, false
	},
	// OpenAI: require tail >= 45 in safe mode
	"openai_api_key": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "sk-") && len(m) >= 3+45 {
			return f, true
		}
		return f, false
	},
	// Slack webhook: ensure host and path shape parse correctly
	"slack_webhook": func(f types.Finding) (types.Finding, bool) {
		u, err := url.Parse(f.Match)
		if err != nil {
			return f, false
		}
		if u.Host != "hooks.slack.com" {
			return f, false
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) != 4 { // services/T/B/token
			return f, false
		}
		return f, true
	},
	// Discord webhook: ensure host and basic path parts
	"discord_webhook": func(f types.Finding) (types.Finding, bool) {
		u, err := url.Parse(f.Match)
		if err != nil {
			return f, false
		}
		if u.Host != "discord.com" {
			return f, false
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) < 4 || parts[0] != "api" || parts[1] != "webhooks" {
			return f, false
		}
		return f, true
	},
	// GitHub: enforce exact tail length in safe mode
	"github_token": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if (strings.HasPrefix(m, "ghp_") || strings.HasPrefix(m, "gho_") || strings.HasPrefix(m, "ghu_") || strings.HasPrefix(m, "ghs_") || strings.HasPrefix(m, "ghr_")) && len(m) == 4+36 {
			return f, true
		}
		return f, false
	},
	// AWS Access Key: enforce AKIA/ASIA + 16 uppercase alnum
	"aws_access_key": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if (strings.HasPrefix(m, "AKIA") || strings.HasPrefix(m, "ASIA")) && len(m) == 20 {
			return f, true
		}
		return f, false
	},
	// AWS Secret Key: require base64 decode to succeed
	"aws_secret_key": func(f types.Finding) (types.Finding, bool) {
		if _, err := base64.StdEncoding.DecodeString(f.Match); err == nil {
			return f, true
		}
		if _, err := base64.RawStdEncoding.DecodeString(f.Match); err == nil {
			return f, true
		}
		return f, false
	},
	// Slack token: enforce xox* prefix and minimum length
	"slack_token": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "xox") && len(m) >= 15 {
			return f, true
		}
		return f, false
	},
	// Netlify token: nf_ + >= 26 base62 in safe mode
	"netlify_token": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "nf_") && len(m) >= len("nf_")+26 && v.IsAlphabet(m[len("nf_"):], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			return f, true
		}
		return f, false
	},
	// Terraform Cloud: tfe./tfc. + >=34 base62
	"terraform_cloud_token": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if (strings.HasPrefix(m, "tfe.") || strings.HasPrefix(m, "tfc.")) && len(m) >= 4+34 && v.IsAlphabet(m[4:], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			return f, true
		}
		return f, false
	},
	// DigitalOcean PAT: dop_v1_ + 64 hex
	"digitalocean_pat": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "dop_v1_") && len(m) == len("dop_v1_")+64 {
			_, err := hex.DecodeString(m[len("dop_v1_"):])
			return f, err == nil
		}
		return f, false
	},
	// DockerHub PAT: dckr_pat_ + 64 base62
	"dockerhub_pat": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "dckr_pat_") && len(m) == len("dckr_pat_")+64 && v.IsAlphabet(m[len("dckr_pat_"):], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			return f, true
		}
		return f, false
	},
	// SendGrid: SG.<16>.<32+>
	"sendgrid_api_key": func(f types.Finding) (types.Finding, bool) {
		parts := strings.Split(f.Match, ".")
		if len(parts) == 3 && len(parts[1]) == 16 && len(parts[2]) >= 32 {
			return f, true
		}
		return f, false
	},
	// Twilio Auth Token: 32 hex
	"twilio_auth_token": func(f types.Finding) (types.Finding, bool) {
		if len(f.Match) == 32 {
			_, err := hex.DecodeString(f.Match)
			return f, err == nil
		}
		return f, false
	},
	// Okta: SSWS <token> with min length 50
	"okta_api_token": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "SSWS ") && len(m) >= len("SSWS ")+50 {
			return f, true
		}
		return f, false
	},
	// NewRelic: known prefixes and length >=30
	"newrelic_api_key": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if (strings.HasPrefix(m, "NRAK-") || strings.HasPrefix(m, "NRAL-") || strings.HasPrefix(m, "NRII-") || strings.HasPrefix(m, "NRAA-")) && len(m) >= 5+27 {
			return f, true
		}
		return f, false
	},
	// Stripe webhook secret: whsec_ + >= 24 base62
	"stripe_webhook_secret": func(f types.Finding) (types.Finding, bool) {
		m := f.Match
		if strings.HasPrefix(m, "whsec_") && len(m) >= len("whsec_")+24 && v.IsAlphabet(m[len("whsec_"):], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			return f, true
		}
		return f, false
	},
}

// verifyFindings applies optional soft verification in safe mode.
func verifyFindings(fs []types.Finding) []types.Finding {
	if VerifyMode != "safe" || len(fs) == 0 {
		return fs
	}
	out := make([]types.Finding, 0, len(fs))
	for _, f := range fs {
		if vf, ok := verifyFilters[f.Detector]; ok {
			if nf, ok2 := vf(f); ok2 {
				out = append(out, nf)
			}
			continue
		}
		out = append(out, f)
	}
	return out
}
