package inbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antham/yogo/internal/client"
	"github.com/antham/yogo/internal/inbox/internal/mail"
	"github.com/fatih/color"
)

type Render interface {
	Coloured() (string, error)
	JSON() (string, error)
}

const noDataToDisplayMsg = "[no data to display]"
const itemNumber = 15

// Inbox represents a mail collection
type Inbox[M client.MailDoc] struct {
	Name       string      `json:"name"`
	InboxItems []InboxItem `json:"mails"`
	client     client.Client[M]
}

type Sender struct {
	Mail string `json:"mail,omitempty"`
	Name string `json:"name,omitempty"`
}

// Inbox represents a mail sumup in an inbox
type InboxItem struct {
	ID     string     `json:"id"`
	Sender *Sender    `json:"sender,omitempty"`
	Title  string     `json:"title"`
	Date   *time.Time `json:"date,omitempty"`
	Body   string     `json:"body,omitempty"`
	IsSPAM bool       `json:"isSPAM"`
}

// NewInbox creates a new mail inbox
func NewInbox[M client.MailDoc](name string) (*Inbox[M], error) {
	client, err := client.New[M]()
	return &Inbox[M]{
		client:     client,
		Name:       name,
		InboxItems: []InboxItem{},
	}, err
}

// Fetch retrieves the full email content from the given
// inbox email offset
func (i *Inbox[M]) Fetch(offset int) (Render, error) {
	ID := &i.InboxItems[offset].ID
	doc, err := i.client.GetMailPage(i.Name, *ID)
	if err != nil {
		return nil, err
	}
	m := mail.Parse(doc)
	m.ID = *ID
	return &m, nil
}

// Count returns total number of mails available in inbox
func (i *Inbox[M]) Count() int {
	return len(i.InboxItems)
}

// Shrink reduces mails size to given value
func (i *Inbox[M]) Shrink(limit int) {
	if len(i.InboxItems) < limit {
		return
	}

	i.InboxItems = i.InboxItems[:limit]
}

// Add appends a mail to mail list
func (i *Inbox[M]) Add(inboxItem InboxItem) {
	i.InboxItems = append(i.InboxItems, inboxItem)
}

// Delete an email
func (i *Inbox[M]) Delete(position int) error {
	mail := i.InboxItems[position]
	if err := i.client.DeleteMail(i.Name, mail.ID); err != nil {
		return err
	}

	i.InboxItems = append(i.InboxItems[:position], i.InboxItems[position+1:]...)
	return nil
}

// Flush empties an inbox
func (i *Inbox[M]) Flush() error {
	if len(i.InboxItems) == 0 {
		return nil
	}

	if err := i.client.FlushMail(i.Name, i.InboxItems[0].ID); err != nil {
		return err
	}

	i.InboxItems = []InboxItem{}
	return nil
}

func (i *Inbox[M]) GetMails() []InboxItem {
	return i.InboxItems
}

func (i *Inbox[M]) Coloured() (string, error) {
	if i.Count() == 0 {
		return "", errors.New("inbox is empty")
	}

	output := ""
	for index, mail := range i.GetMails() {
		info := struct {
			Index         string
			SenderName    string
			HasSenderName bool
			SenderMail    string
			HasSenderMail bool
			Title         string
			TitlePadding  string
			SPAM          string
		}{}

		if mail.Sender != nil {
			if mail.Sender.Name != "" {
				info.HasSenderName = true
				info.SenderName = color.YellowString(mail.Sender.Name)
			} else {
				info.SenderName = color.YellowString(noDataToDisplayMsg)
			}
			if mail.Sender.Mail != "" {
				info.HasSenderMail = true
				info.SenderMail = color.YellowString(mail.Sender.Mail)
			} else {
				info.SenderMail = color.YellowString(noDataToDisplayMsg)
			}
		} else {
			info.SenderName = color.YellowString(noDataToDisplayMsg)
			info.SenderMail = color.YellowString(noDataToDisplayMsg)
		}
		if mail.Title != "" {
			info.Title = color.CyanString(mail.Title)
		} else {
			info.Title = color.CyanString(noDataToDisplayMsg)
		}
		if mail.IsSPAM {
			info.SPAM = color.RedString("[SPAM]")
		}
		info.Index = strconv.Itoa(index + 1)

		for i := 0; i < len(info.Index); i++ {
			info.TitlePadding = info.TitlePadding + " "
		}

		var buf bytes.Buffer
		tpl := template.Must(template.New("t").Parse(` {{.Index}} {{ if .HasSenderName -}}
{{- .SenderName -}}
{{- end -}}
{{- if (and .HasSenderMail .HasSenderName) }} {{ end -}}
{{- if (and (eq .HasSenderMail false) (eq .HasSenderName false)) }}{{ .SenderName }}{{- end -}}
{{- if .HasSenderMail -}}
	{{- if .HasSenderName -}}<{{- end -}}
	{{- .SenderMail -}}
	{{- if .HasSenderName -}}>{{- end -}}
{{- end -}}
{{- if .SPAM }} {{ .SPAM -}}{{- end -}}
{{- if .Title }}
  {{.TitlePadding}}{{ .Title }}
{{ end }}
`))
		if err := tpl.Execute(&buf, info); err != nil {
			return "", err
		}
		output = output + buf.String()
	}
	return strings.TrimRight(output, "\n"), nil
}

func (i *Inbox[M]) JSON() (string, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return "", errors.New("something wrong occurred")
	}
	s := string(data)
	return s, nil
}

// ParseInboxPages parses inbox email in given page
func (i *Inbox[M]) ParseInboxPages(limit int) error {
	for page := 1; page <= (limit/itemNumber)+1 && limit >= i.Count(); page++ {
		doc, err := i.client.GetMailsPage(i.Name, page)
		if err != nil {
			return err
		}

		parseInboxPage(doc, i)
		time.Sleep(1 * time.Second)
	}

	i.Shrink(limit)

	return nil
}

// ParseInboxPage parses inbox email in given page
func parseInboxPage[M client.MailDoc](doc *goquery.Document, inbox *Inbox[M]) {
	doc.Find("div.m").Each(func(i int, s *goquery.Selection) {
		var isSPAM bool
		name := s.Find("span.lmf").Text()
		userEmail := name

		if len(name) >= 6 && name[:6] == "[SPAM]" {
			isSPAM = true
			name = name[6:]
		}

		if strings.Contains(name, "@") {
			name = ""
		} else {
			userEmail = ""
		}

		if ID, ok := s.Attr("id"); ok {
			inboxItem := InboxItem{
				ID: ID,
				Sender: &Sender{
					Name: name,
					Mail: userEmail,
				},
				Title:  s.Find("div.lms").Text(),
				IsSPAM: isSPAM,
			}

			inbox.Add(inboxItem)
		}
	})
}
