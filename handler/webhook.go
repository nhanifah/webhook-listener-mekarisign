package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"webhook-listener-mekarisign/config"
	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/logger"
	"webhook-listener-mekarisign/model"
	"webhook-listener-mekarisign/service"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	db       *database.Database
	xs       *service.XenditService
	cfg      *config.Config
	rabbitMQ *service.RabbitMQService
}

type InvoiceData struct {
	Amount         int `json:"amount"`
	AvailableBanks []struct {
		AccountHolderName string `json:"account_holder_name"`
		BankBranch        string `json:"bank_branch"`
		BankCode          string `json:"bank_code"`
		CollectionType    string `json:"collection_type"`
		TransferAmount    int    `json:"transfer_amount"`
	} `json:"available_banks"`
	AvailableDirectDebits []struct {
		DirectDebitType string `json:"direct_debit_type"`
	} `json:"available_direct_debits"`
	AvailableEwallets []struct {
		EwalletType string `json:"ewallet_type"`
	} `json:"available_ewallets"`
	AvailablePaylaters []struct {
		PaylaterType string `json:"paylater_type"`
	} `json:"available_paylaters"`
	AvailableQrCodes []struct {
		QrCodeType string `json:"qr_code_type"`
	} `json:"available_qr_codes"`
	AvailableRetailOutlets []struct {
		RetailOutletName string `json:"retail_outlet_name"`
	} `json:"available_retail_outlets"`
	Created                   time.Time `json:"created"`
	Currency                  string    `json:"currency"`
	Description               string    `json:"description"`
	ExpiryDate                time.Time `json:"expiry_date"`
	ExternalID                string    `json:"external_id"`
	ID                        string    `json:"id"`
	InvoiceURL                string    `json:"invoice_url"`
	MerchantName              string    `json:"merchant_name"`
	MerchantProfilePictureURL string    `json:"merchant_profile_picture_url"`
	PayerEmail                string    `json:"payer_email"`
	ShouldExcludeCreditCard   bool      `json:"should_exclude_credit_card"`
	ShouldSendEmail           bool      `json:"should_send_email"`
	Status                    string    `json:"status"`
	Updated                   time.Time `json:"updated"`
	UserID                    string    `json:"user_id"`
}

type Signer struct {
	Email  string `bson:"email" json:"email"`
	Name   string `bson:"name" json:"name"`
	Order  int    `bson:"order" json:"order"`
	Status string `bson:"status" json:"status"`
	Phone  string `bson:"phone" json:"phone"`
}

var mekariSignData struct {
	Data struct {
		Attributes struct {
			Signers []Signer `bson:"signers"`
		} `bson:"attributes"`
	} `bson:"data"`
}

func NewWebhookHandler(db *database.Database, xs *service.XenditService, cfg *config.Config, rabbitMQ *service.RabbitMQService) *WebhookHandler {
	return &WebhookHandler{db: db, xs: xs, cfg: cfg, rabbitMQ: rabbitMQ}
}

func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	var req map[string]interface{}

	// Parsing JSON request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	logger.NewLogger().Info("Webhook received", zap.Any("request", req))

	// ensure _id is an ObjectID
	if req["_id"] == nil {
		req["_id"] = primitive.NewObjectID()
	} else {
		if _, ok := req["_id"].(primitive.ObjectID); !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "_id must be an ObjectID"})
		}
	}

	collectionName := h.getCollectionName(req)
	collection := h.db.DB.Collection(collectionName)

	// validation untuk xendit webhook
	if collectionName == "webhook_xendit" {
		token := c.Request().Header.Get("x-callback-token")
		if token == "" || token != h.cfg.XenditCallbackToken {
			logger.NewLogger().Warn("Unauthorized Xendit webhook request", zap.String("received_token", token))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
	}

	// Simpan ke MongoDB
	filter := bson.M{"_id": req["_id"]}
	update := bson.M{"$set": req}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		logger.NewLogger().Error("Failed to save to database", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save to database"})
	}

	if collectionName == "webhook_mekarisign" {
		return h.handleMekariSignWebhook(req, res.UpsertedID, c)
	} else if collectionName == "webhook_xendit" {
		return h.handleXenditWebhook(req, c)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Webhook received"})
}

func (h *WebhookHandler) getCollectionName(req map[string]interface{}) string {
	if req["external_id"] != nil {
		return "webhook_xendit"
	} else if req["data"] != nil {
		return "webhook_mekarisign"
	}
	return "webhook_logs"
}

func (h *WebhookHandler) handleMekariSignWebhook(req map[string]interface{}, insertedID interface{}, c echo.Context) error {
	data, ok := req["data"].(map[string]interface{})
	if !ok {
		return c.JSON(http.StatusOK, bson.M{"message": "Webhook received but no invoice created", "inserted_id": insertedID})
	}

	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid attributes structure"})
	}

	signingStatus, _ := attributes["signing_status"].(string)
	stampingStatus, _ := attributes["stamping_status"].(string)
	docUrl, _ := attributes["doc_url"].(string)
	docId, _ := data["id"].(string)
	fileName, _ := attributes["filename"].(string)

	// push to student attachment
	// get signer data
	signers, _ := attributes["signers"].([]interface{})
	for i := 0; i < len(signers); i++ {
		signer, _ := signers[i].(map[string]interface{})
		if signer["email"] == h.cfg.DirectorMail {
			signers = append(signers[:i], signers[i+1:]...)
			i--
		}
	}

	if len(signers) > 0 {
		payerEmail, _ := signers[0].(map[string]interface{})["email"].(string)
		// get student data
		student, err := model.GetStudentByEmail(payerEmail)
		if student == nil {
			logger.NewLogger().Error("pushToStudentAttachment => Failed to get students", zap.Error(err))
		}

		var studentID sql.NullString
		if student != nil {
			studentID = sql.NullString{String: student.ID, Valid: true}
		} else {
			studentID = sql.NullString{Valid: false}
		}

		now := time.Now()

		// logger.NewLogger().Info("pushToStudentAttachment", zap.String("docId", docId))
		// logger.NewLogger().Info("pushToStudentAttachment", zap.String("docUrl", docUrl))
		// logger.NewLogger().Info("pushToStudentAttachment", zap.String("fileName", fileName))
		// logger.NewLogger().Info("pushToStudentAttachment", zap.String("payerEmail", payerEmail))
		// logger.NewLogger().Info("pushToStudentAttachment", zap.String("studentID", studentID.String))

		// create student attachment
		studentAttachment := model.StudentAttachment{
			ID:         docId,
			StudentID:  studentID,
			FileName:   fileName,
			FileURL:    "https://api.mekari.com" + docUrl,
			UploadedAt: sql.NullTime{Time: now, Valid: true},
			CreatedAt:  sql.NullTime{Time: now, Valid: true},
			UpdatedAt:  sql.NullTime{Time: now, Valid: true},
			DeletedAt:  sql.NullTime{Valid: false},
		}

		// save to database
		_, err = model.CreateOrUpdateStudentAttachment(studentAttachment)
		if err != nil {
			logger.NewLogger().Error("pushToStudentAttachment => Failed to save student attachment", zap.Error(err))
		}
	}

	if signingStatus == "completed" && stampingStatus == "success" {
		if len(signers) > 0 {
			return h.createInvoiceForMekariSign(signers[0], data["id"].(string), 1, c)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Webhook received but no invoice created"})
}

func (h *WebhookHandler) createInvoiceForMekariSign(signer interface{}, dataID string, paymentFor int, c echo.Context) error {

	// data dump for signer
	logger.NewLogger().Info("createInvoiceForMekariSign", zap.Any("signer", signer))
	logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("dataID", dataID))
	logger.NewLogger().Info("createInvoiceForMekariSign", zap.Int("paymentFor", paymentFor))

	var payerPhone string
	signerData, dataExists := signer.(map[string]interface{})
	payerEmail, emailExists := signerData["email"].(string)
	payerName, nameExists := signerData["name"].(string)
	payerPhone, phoneExists := signerData["phone"].(string)

	// logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("payerEmail", payerEmail))
	// logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("payerName", payerName))
	// logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("payerPhone", payerPhone))
	// logger.NewLogger().Info("createInvoiceForMekariSign", zap.Bool("payerPhone exist", phoneExists))

	if !dataExists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Signer data is missing"})
	}
	if phoneExists && payerPhone != "" {
		logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("Sanitizing phone:", payerPhone))
		// return c.JSON(http.StatusBadRequest, map[string]string{"error": "Phone number is missing"})
		// Remove + if start with +
		if payerPhone[0] == '+' {
			payerPhone = payerPhone[1:]
		}
		if payerPhone[:2] == "08" {
			payerPhone = "628" + payerPhone[2:]
		}

		payerPhone = fmt.Sprintf("+%s", payerPhone)
	} else {
		payerPhone = "+"
	}
	if !nameExists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is missing"})
	}
	if !emailExists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is missing"})
	}

	externalID := fmt.Sprintf("q%d-goglobal-%s", paymentFor, dataID)
	logger.NewLogger().Info("Creating invoice for MekariSign webhook", zap.String("external_id", externalID))

	customer := &service.CustomerObject{
		Id:           externalID,
		PhoneNumber:  payerPhone,
		GivenNames:   payerName,
		Surname:      " ",
		Email:        payerEmail,
		MobileNumber: payerPhone,
		CustomerId:   externalID,
	}

	logger.NewLogger().Info("Customer object", zap.Any("customer", customer))

	description := fmt.Sprintf("Pembayaran ke-%d Go Global Indonesia kepada %s", paymentFor, payerName)

	invDuration := h.cfg.XenditInvDuration
	if paymentFor == 2 {
		// set expiry date to date 5 april 2025 for second payment
		// count day from now to 5 april 2025
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			logger.NewLogger().Error("Failed to load location", zap.Error(err))
		}

		// Set expiry date in Jakarta time
		expiryDate := time.Date(2025, 4, 30, 23, 59, 0, 0, loc)
		invDuration = fmt.Sprintf("%.0f", time.Until(expiryDate).Seconds())
	}

	logger.NewLogger().Info("createInvoiceForMekariSign", zap.String("Invoice duration", invDuration))

	// check data from db
	student, err := model.GetStudentByEmail(payerEmail)
	if student == nil {
		logger.NewLogger().Error("createInvoiceForMekariSign => Failed to get students", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get students"})
	}

	// get class data
	var amount float64
	if student.ProgramType.Valid { // Pastikan ProgramType tidak NULL
		switch student.ProgramType.String {
		case "kelas_senin_jumat_siang":
			amount = 6250000
		case "kelas_senin_jumat_malam":
			amount = 6000000
		case "kelas_akhir_pekan":
			amount = 6000000
		default:
			amount = 7500000 // Atur default jika tidak cocok
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ProgramType is NULL"})
		}
	} else {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get program type"})
	}

	logger.NewLogger().Info("Amount", zap.Float64("amount", amount))
	logger.NewLogger().Info("Student", zap.Any("student", student))

	invoice, err := h.xs.CreateInvoice(
		externalID,
		payerEmail,
		description,
		customer,
		invDuration,
		amount,
		"",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invoice"})
	}

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal invoice"})
	}

	var invoiceResponse InvoiceData
	if err := json.Unmarshal(invoiceJSON, &invoiceResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse invoice response"})
	}

	logger.NewLogger().Info("Invoice created", zap.Any("invoice", invoiceResponse))

	// Publish invoice to RabbitMQ
	emailData := service.EmailSendStruct{
		To:            payerEmail,
		Subject:       description,
		RecipientName: payerName,
		Link:          invoiceResponse.InvoiceURL,
		DueDate:       "beberapa hari kedepan",
	}

	if paymentFor == 1 {
		templateName := "template_send_email_sign_success.html"
		err = h.rabbitMQ.SendEmail(emailData.To, emailData.Subject, templateName, emailData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
		}
	} else if paymentFor == 2 {
		templateName := "template_send_email_payment_1_success.html"
		err = h.rabbitMQ.SendEmail(emailData.To, emailData.Subject, templateName, emailData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Webhook processed and invoice created",
		"invoice_link": invoiceResponse.InvoiceURL,
		"invoice_id":   invoiceResponse.ID,
	})
}

func (h *WebhookHandler) handleXenditWebhook(req map[string]interface{}, c echo.Context) error {
	logger.NewLogger().Info("Xendit webhook received", zap.Any("request", req))
	xenditDesc, ok := req["description"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid description"})
	}
	xenditStatus, ok := req["status"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid status"})
	}

	// jika xenditDesc mengandung kata "Pembayaran ke-1" dan xenditStatus adalah "PAID" maka akan di proses pembuatan invoice baru dengan nama Pembayaran ke-2
	if matched, _ := regexp.MatchString(`Pembayaran ke-1`, xenditDesc); matched && xenditStatus == "PAID" {
		// send payment notification email
		payerNotificationEmail, _ := req["payer_email"].(string)
		payerNotificationDesc, _ := req["description"].(string)
		payerNotificationInvoice, _ := req["id"].(string)
		payerNotificationName := req["description"].(string)[43:]

		emailNotificationData := service.PaymentNotificationStruct{
			To:            payerNotificationEmail,
			Subject:       payerNotificationDesc,
			RecipientName: payerNotificationName,
			InvoiceID:     "https://checkout.xendit.co/web/" + payerNotificationInvoice,
			Amount:        fmt.Sprintf("Rp %s", humanize.Comma(int64(req["amount"].(float64)))),
		}

		err := h.rabbitMQ.SendPaymentNotification("aliffirdi07@gmail.com", "Pendaftaran LPK Go Global", "template_send_email_success_payment.html", emailNotificationData)
		if err != nil {
			logger.NewLogger().Error("Failed to send email", zap.Error(err))
			// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
		}

		// sanitize "q1-goglobal-<data_id>" menjadi "<data_id>"
		dataID := req["external_id"].(string)[12:]

		logger.NewLogger().Info("Creating invoice for Xendit webhook", zap.String("external_id", dataID))

		// Get data from 'webhook_mekarisign' collection
		mekariSignCollection := h.db.DB.Collection("webhook_mekarisign")
		if err := mekariSignCollection.FindOne(context.Background(), bson.M{"data.id": dataID}).Decode(&mekariSignData); err != nil {
			logger.NewLogger().Error("Failed to get data from MekariSign collection", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get data from MekariSign collection"})
		}

		signers := mekariSignData.Data.Attributes.Signers

		// Filter out Email "angga@go-global.id"
		filteredSigners := []Signer{}
		for _, signer := range signers {
			if signer.Email != h.cfg.DirectorMail {
				filteredSigners = append(filteredSigners, signer)
			}
		}

		logger.NewLogger().Info("Filtered signers", zap.Any("signers", filteredSigners))

		// return for debug purpose
		// return c.JSON(http.StatusOK, map[string]interface{}{
		// 	"message": "Webhook processed and invoice created",
		// 	"signers": b,
		// })

		// convert filteredSigners[0] to map[string]interface{}
		// filteredSigners[0] is a struct, so we need to convert it to map[string]interface{}
		// to be able to pass it to createInvoiceForMekariSign
		signerJSON, err := json.Marshal(filteredSigners[0])
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal signer"})
		}

		var signerMap map[string]interface{}
		if err := json.Unmarshal(signerJSON, &signerMap); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse signer"})
		}

		payerPhone := signerMap["phone"].(string)

		// Send wa notification
		whatsappPayload := service.WhatsAppPayload{
			Sender:          "WKWK JAPANESE",
			Recipient:       h.cfg.WhatsappNotificationNumber,
			RecipientName:   payerNotificationName,
			TemplateID:      "194cd576-8d34-4f5f-b9ca-29e97d1bbe90",
			Attrb:           []string{"Pembayaran Kedua", "Program Basic Kelas A", "Batch 1.0", fmt.Sprintf("Rp %s", humanize.Comma(int64(req["amount"].(float64)))), payerNotificationName, payerNotificationEmail, payerPhone, payerNotificationInvoice},
			EnabledSchedule: 0,
		}

		err = service.SendWhatsAppMessage(h.cfg, whatsappPayload)
		if err != nil {
			logger.NewLogger().Error("Failed to send WhatsApp message", zap.Error(err))
		}

		// create invoice for filteredSigners[0]
		logger.NewLogger().Info("Creating invoice for filtered signer", zap.Any("signer", signerMap))

		return h.createInvoiceForMekariSign(signerMap, dataID, 2, c)
	}

	// jika xenditDesc mengandung kata "Pembayaran ke-2" dan xenditStatus adalah "PAID" maka akan di proses pembuatan invoice baru dengan nama Pembayaran ke-2
	if matched, _ := regexp.MatchString(`Pembayaran ke-2`, xenditDesc); matched && xenditStatus == "PAID" {
		// send payment notification email
		payerNotificationEmail, _ := req["payer_email"].(string)
		payerNotificationDesc, _ := req["description"].(string)
		payerNotificationInvoice, _ := req["id"].(string)
		payerNotificationName := req["description"].(string)[43:]

		emailNotificationData := service.PaymentNotificationStruct{
			To:            payerNotificationEmail,
			Subject:       payerNotificationDesc,
			RecipientName: payerNotificationName,
			InvoiceID:     "https://checkout.xendit.co/web/" + payerNotificationInvoice,
			Amount:        fmt.Sprintf("Rp %s", humanize.Comma(int64(req["amount"].(float64)))),
		}

		err := h.rabbitMQ.SendPaymentNotification("aliffirdi07@gmail.com", "Pendaftaran LPK Go Global", "template_send_email_success_payment.html", emailNotificationData)
		if err != nil {
			logger.NewLogger().Error("Failed to send email", zap.Error(err))
			// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
		}

		// sanitize "q1-goglobal-<data_id>" menjadi "<data_id>"
		dataID := req["external_id"].(string)[12:]

		logger.NewLogger().Info("Creating invoice for Xendit webhook", zap.String("external_id", dataID))

		// Get data from 'webhook_mekarisign' collection
		mekariSignCollection := h.db.DB.Collection("webhook_mekarisign")
		if err := mekariSignCollection.FindOne(context.Background(), bson.M{"data.id": dataID}).Decode(&mekariSignData); err != nil {
			logger.NewLogger().Error("Failed to get data from MekariSign collection", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get data from MekariSign collection"})
		}

		signers := mekariSignData.Data.Attributes.Signers

		// Filter out Email "angga@go-global.id"
		filteredSigners := []Signer{}
		for _, signer := range signers {
			if signer.Email != h.cfg.DirectorMail {
				filteredSigners = append(filteredSigners, signer)
			}
		}

		logger.NewLogger().Info("Filtered signers", zap.Any("signers", filteredSigners))

		payerEmail := filteredSigners[0].Email
		payerName := filteredSigners[0].Name
		payerPhone := filteredSigners[0].Phone

		if payerPhone[0] == '+' {
			payerPhone = payerPhone[1:]
		}
		if payerPhone[:2] == "08" {
			payerPhone = "628" + payerPhone[2:]
		}

		emailData := service.EmailSendStruct{
			To:            payerEmail,
			Subject:       "Pembayaran ke-2 Telah Lunas",
			RecipientName: payerName,
			Link:          "",
			DueDate:       "beberapa hari kedepan",
		}
		templateName := "template_send_email_payment_2_success.html"
		err = h.rabbitMQ.SendEmail(emailData.To, emailData.Subject, templateName, emailData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email"})
		}
		// Send wa notification
		whatsappPayload := service.WhatsAppPayload{
			Sender:          "WKWK JAPANESE",
			Recipient:       h.cfg.WhatsappNotificationNumber,
			RecipientName:   payerName,
			TemplateID:      "194cd576-8d34-4f5f-b9ca-29e97d1bbe90",
			Attrb:           []string{"Pembayaran Kedua", "Program Basic Kelas A", "Batch 1.0", fmt.Sprintf("Rp %s", humanize.Comma(int64(req["amount"].(float64)))), payerNotificationName, payerNotificationEmail, payerPhone, payerNotificationInvoice},
			EnabledSchedule: 0,
		}

		err = service.SendWhatsAppMessage(h.cfg, whatsappPayload)
		if err != nil {
			logger.NewLogger().Error("Failed to send WhatsApp message", zap.Error(err))
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Webhook received"})
}
