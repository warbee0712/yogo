package inbox

import (
	"io/ioutil"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/mail?b=test&id=me_ZwRjAwRkZwRkAQV1ZQNjBGD4AGL4AD%3D%3D",
			"mail.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	parseInboxPage(getDoc(t, "inbox_page.html"), inbox)

	assert.Equal(t, 15, inbox.Count())
	m, err := inbox.Fetch(0)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRkZwRkAQV1ZQNjBGD4AGL4AD==", m.ID)
	assert.Equal(t, "In any case, I am happy that we met", m.Title)
}

func TestCount(t *testing.T) {
	inbox := &Inbox{}

	parseInboxPage(getDoc(t, "inbox_page.html"), inbox)

	assert.Equal(t, inbox.Count(), 15)
}

func TestParseInboxPages(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=2&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_2.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/en/mail?b=test&id=me_ZwRjAwRmZGtmZQR0ZQNjAwt2BGV5BN%3D%3D",
			"mail.html",
		},
		{
			"GET",
			"https://yopmail.com/en/mail?b=test&id=me_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj%3D%3D",
			"mail.html",
		},
		{
			"GET",
			"https://yopmail.com/en/mail?b=test&id=me_ZwRjAwRmZGtmZwR0ZQNjAwt3AmxlZN%3D%3D",
			"mail.html",
		},
		{
			"GET",
			"https://yopmail.com/en/mail?b=test&id=me_ZwRjAwRmZGtmZwN3ZQNjAwt3AmZlAD%3D%3D",
			"mail2.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(29)

	assert.NoError(t, err)
	assert.Equal(t, "test", inbox.Name)
	assert.Equal(t, 29, inbox.Count())
	m, err := inbox.Fetch(0)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj==", m.ID)
	m, err = inbox.Fetch(28)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRmZGtmZQR0ZQNjAwt2BGV5BN==", m.ID)
	m, err = inbox.Fetch(13)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRmZGtmZwR0ZQNjAwt3AmxlZN==", m.ID)	
	assert.False(t, m.IsSPAM)
	assert.Equal(t, "In any case, I am happy that we met", m.Title)
	m, err = inbox.Fetch(14)
	assert.Equal(t, "e_ZwRjAwRmZGtmZwN3ZQNjAwt3AmZlAD==", m.ID)		
	assert.NoError(t, err)
	assert.False(t, m.IsSPAM)
	assert.Equal(t, `A title`, m.Title)
}

func TestShrink(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=2&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_2.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(19)

	assert.NoError(t, err)
	assert.Equal(t, 19, inbox.Count())
	m, err := inbox.Fetch(0)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj==", m.ID)
	m, err = inbox.Fetch(18)
	assert.NoError(t, err)
	assert.Equal(t, "e_ZwRjAwRmZGtmZGDlZQNjAwt3AGHkAt==", m.ID)
}

func TestShrinkEmptyInbox(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_empty.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(1)

	assert.NoError(t, err)
	assert.Equal(t, 0, inbox.Count())
}

func TestShrinkWithLimitGreaterThanNumberOfMessagesAvailable(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=2&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_empty.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(18)

	assert.NoError(t, err)
	assert.Equal(t, 15, inbox.Count())
}

func TestGetAll(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=2&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_2.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(29)
	mails := inbox.InboxItems

	assert.NoError(t, err)
	assert.Len(t, mails, 29)
	assert.Equal(t, "e_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj==", mails[0].ID)
	assert.Equal(t, "e_ZwRjAwRmZGtmZQR0ZQNjAwt2BGV5BN==", mails[28].ID)
}

func TestFlush(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=2&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_2.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?login=test&p=1&d=all&ctrl=e_ZGtkZwVmZQNmBGV1ZQNjZQVjAwD1BD==&v=4.8&r_c=&id",
			"noop.html",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(15)
	inbox.Flush()

	assert.NoError(t, err)
}

func TestFlushEmptyInbox(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_empty.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(1)
	inbox.Flush()

	assert.NoError(t, err)
	assert.Equal(t, 0, inbox.Count())
}

func TestDelete(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assert.NoError(t, registerResponders([]responder{
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=&id=&login=test&p=1&r_c=&scrl=&spam=true&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"inbox_page_1.html",
		},
		{
			"GET",
			"https://yopmail.com",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/ver/4.8/webmail.js",
			"webmail.js",
		},
		{
			"GET",
			"https://yopmail.com/consent?c=accept",
			"main_page.html",
		},
		{
			"GET",
			"https://yopmail.com/en/inbox?ctrl=&d=e_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj%3D%3D&id=&login=test&p=1&r_c=&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp",
			"noop.html",
		},
	}))

	inbox, err := NewInbox("test")
	assert.NoError(t, err)

	err = inbox.ParseInboxPages(1)
	assert.NoError(t, inbox.Delete(0))

	assert.Equal(t, 1, httpmock.GetCallCountInfo()["GET https://yopmail.com/en/inbox?ctrl=&d=e_ZwRjAwRmZGtmAwZ1ZQNjAwt5AQZmZj%3D%3D&id=&login=test&p=1&r_c=&v=4.8&yj=VZGV5AmpjZwp5ZGNmZwL0BQH&yp=UAQDkAGH2Amp2Zmt0ZmVmAGp"])
	assert.NoError(t, err)
}

type responder struct {
	method   string
	URL      string
	filename string
}

func registerResponders(responders []responder) error {
	for _, r := range responders {
		b, err := ioutil.ReadFile(r.filename)
		if err != nil {
			return err
		}

		httpmock.RegisterResponder(r.method, r.URL,
			httpmock.NewStringResponder(200, string(b)))
	}
	return nil
}
