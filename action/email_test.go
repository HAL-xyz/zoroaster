package action

// Actually send an email. Commented out bc we only want
// to run it manually
//func TestSendEmail(t *testing.T) {
//
//	sesSession := aws.GetSESSession()
//	to := []string{"manlio.poltronieri@gmail.com", "marco@atomic.eu.com"}
//	subject := "hello from Zoroaster to both of you :)"
//	body := `
//		I'm testing that emails are sent as text/plain only,
//		with newlines etc.
//		Like this is a new line.
//	`
//	res, err := sendEmail(sesSession, to, subject, body)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Println(res)
//}
