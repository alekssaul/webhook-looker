package main

// lookerWebhook struct allows validation of the webhook
type lookerWebhook struct {
	LookerInstance     string
	LookerWebhookToken string
}

// lookerBody struct describes the Webhook request body
type lookerBody struct {
	ScheduledPlan struct {
		QueryID               string      `json:"query_id"`
		ScheduledPlanID       string      `json:"scheduled_plan_id"`
		DownloadURL           interface{} `json:"download_url"`
		FiltersDifferFromLook bool        `json:"filters_differ_from_look"`
		Title                 string      `json:"title"`
		Query                 interface{} `json:"query"`
		URL                   string      `json:"url"`
		Type                  string      `json:"type"`
	} `json:"scheduled_plan"`
	Attachment struct {
		Extension string `json:"extension"`
		Data      string `json:"data"`
		Mimetype  string `json:"mimetype"`
	} `json:"attachment"`
	Type       string `json:"type"`
	FormParams struct {
	} `json:"form_params"`
	Data interface{} `json:"data"`
}
