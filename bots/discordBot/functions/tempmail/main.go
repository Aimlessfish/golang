package tempmail

import (
	"discordBot/util"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GuerrillaEmailListResponse represents the response from get_email_list
type GuerrillaEmailListResponse struct {
	List []struct {
		MailID        string `json:"mail_id"`
		MailFrom      string `json:"mail_from"`
		MailSubject   string `json:"mail_subject"`
		MailExcerpt   string `json:"mail_excerpt"`
		MailTimestamp string `json:"mail_timestamp"`
	} `json:"list"`
}

// GuerrillaInboxResponse represents the response from Guerrilla Mail inbox
type GuerrillaInboxResponse struct {
	List     []GuerrillaMailItem `json:"list"`
	Email    string              `json:"email"`
	Alias    string              `json:"alias"`
	Ts       string              `json:"ts"`
	SidToken string              `json:"sid_token"`
	Count    string              `json:"count"`
	Users    string              `json:"users"`
	Stats    struct {
		SequenceMail     string `json:"sequence_mail"`
		CreatedAddresses string `json:"created_addresses"`
		ReceivedEmails   string `json:"received_emails"`
		Total            string `json:"total"`
		TotalPerHour     string `json:"total_per_hour"`
	} `json:"stats"`
	Auth struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error_codes"`
	} `json:"auth"`
}

func (g *GuerrillaInboxResponse) UnmarshalJSON(data []byte) error {
	type Alias GuerrillaInboxResponse
	aux := &struct {
		Ts json.RawMessage `json:"ts"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// ts: string or number
	var tsString string
	if err := json.Unmarshal(aux.Ts, &tsString); err == nil {
		g.Ts = tsString
	} else {
		var tsNumber float64
		if err := json.Unmarshal(aux.Ts, &tsNumber); err == nil {
			g.Ts = fmt.Sprintf("%.0f", tsNumber)
		} else {
			g.Ts = ""
		}
	}
	return nil
}

// GuerrillaMailItem handles mail_timestamp as string or number
type GuerrillaMailItem struct {
	MailID        string `json:"mail_id"`
	MailFrom      string `json:"mail_from"`
	MailSubject   string `json:"mail_subject"`
	MailExcerpt   string `json:"mail_excerpt"`
	MailTimestamp string `json:"mail_timestamp"`
}

func (m *GuerrillaMailItem) UnmarshalJSON(data []byte) error {
	type Alias GuerrillaMailItem
	aux := &struct {
		MailID        json.RawMessage `json:"mail_id"`
		MailTimestamp json.RawMessage `json:"mail_timestamp"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// mail_id: string or number
	var idString string
	if err := json.Unmarshal(aux.MailID, &idString); err == nil {
		m.MailID = idString
	} else {
		var idNumber float64
		if err := json.Unmarshal(aux.MailID, &idNumber); err == nil {
			m.MailID = fmt.Sprintf("%.0f", idNumber)
		} else {
			m.MailID = ""
		}
	}
	// mail_timestamp: string or number
	var tsString string
	if err := json.Unmarshal(aux.MailTimestamp, &tsString); err == nil {
		m.MailTimestamp = tsString
	} else {
		var tsNumber float64
		if err := json.Unmarshal(aux.MailTimestamp, &tsNumber); err == nil {
			m.MailTimestamp = fmt.Sprintf("%.0f", tsNumber)
		} else {
			m.MailTimestamp = ""
		}
	}
	// The rest
	m.MailFrom = aux.MailFrom
	m.MailSubject = aux.MailSubject
	m.MailExcerpt = aux.MailExcerpt
	return nil
}

// GetRandomYopmail fetches a random email from yopmail.com/en/email-generator
func GetRandomYopmail() (string, string, error) {
	resp, err := http.Get("https://yopmail.com/en/email-generator")
	if err != nil {
		fmt.Printf("[DEBUG] Failed to fetch yopmail page: %v\n", err)
		return "", "", fmt.Errorf("failed to fetch yopmail page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[DEBUG] Non-200 status code: %d\n", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to parse yopmail HTML: %v\n", err)
		return "", "", fmt.Errorf("failed to parse yopmail HTML: %w", err)
	}

	rawemail := doc.Find("#geny").Text()
	// Clean the email: take only up to the first '@' or the whole string if no '@'
	email := rawemail
	if at := strings.Index(rawemail, "@"); at != -1 {
		email = rawemail[:at]
	}
	alternateDomains, err := getYopAlternateDomains()
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get alternate domains: %v\n", err)
		return "", "", err
	}
	return email, strings.Join(alternateDomains, ", "), nil
}

// GetAlternateDomains fetches alternate domains from yopmail.com/en/alternate-domains
func getYopAlternateDomains() ([]string, error) {
	// Use curl to fetch the page and extract <div> contents using regex
	cmd := "curl -s 'https://yopmail.com/en/domain?d=all'"
	output, err := util.ExecCommandOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alternate domains page: %w", err)
	}

	// Use regex to extract all <div>...</div> contents
	re := regexp.MustCompile(`<div[^>]*>(.*?)</div>`) // non-greedy match
	matches := re.FindAllStringSubmatch(output, -1)
	var domains []string
	for _, match := range matches {
		if len(domains) >= 10 {
			break
		}
		domain := strings.TrimSpace(match[1])
		// Skip if it looks like a tag or is empty
		if domain == "" || strings.HasPrefix(domain, "<") {
			continue
		}
		domains = append(domains, domain)
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("could not find alternate domains in page")
	}
	return domains, nil
}

// geurrilla temp mail
func GetRandomGuerrillaEmail() (string, string, error) {
	util.LoggerInit("tempmail", "GetGuerrillaMail")
	cmd := "curl -s https://api.guerrillamail.com/ajax.php?f=get_email_address"
	output, err := util.ExecCommandOutput(cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch guerrilla mail: %w", err)
	}

	var resp map[string]interface{}
	err = json.Unmarshal([]byte(output), &resp)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse guerrilla mail response: %w", err)
	}
	err = geurrillaSetAddress(resp["sid_token"].(string))
	if err != nil {
		return "", "", fmt.Errorf("failed to set guerrilla email address: %w", err)
	}

	return resp["email_addr"].(string), resp["sid_token"].(string), nil
}

// GetGuerrillaInboxRaw fetches the raw JSON inbox string from Guerrilla Mail API
func GetGuerrillaInboxRaw(sidToken string) (string, error) {
	logger := util.LoggerInit("tempmail", "GetGuerrillaInboxRaw")
	url := "https://api.guerrillamail.com/ajax.php?f=get_email_list&offset=0&sid_token=" + sidToken
	logger.Info(url)
	output, err := util.ExecCommandOutput("curl -s '" + url + "'")
	if err != nil {
		return "", fmt.Errorf("failed to fetch guerrilla inbox list: %w", err)
	}
	return output, nil
}

func geurrillaSetAddress(uid string) error {
	logger := util.LoggerInit("tempmail", "geurrillaSetAddress")
	cmd := fmt.Sprintf("curl -s 'https://api.guerrillamail.com/ajax.php?f=set_email_user&email_user=%s&lang=en&sid_token='", uid)
	output, err := util.ExecCommandOutput(cmd)
	if err != nil {
		logger.Error("failed to set guerrilla email address:", "error", err, "output<%s>", output)
		return fmt.Errorf("failed to set guerrilla email address: %w", err)
	}
	return nil
}
