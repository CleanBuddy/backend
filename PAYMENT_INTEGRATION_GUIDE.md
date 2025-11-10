# Payment Integration Guide

## Overview

CleanBuddy uses **Netopia Payments** as the exclusive payment provider - Romania's leading payment processor with the highest market share and reliability.

Netopia supports all major card types (Visa, Mastercard, Maestro) and provides preauthorization, capture, and refund capabilities required for the CleanBuddy booking flow.

---

## Payment Flow

### 1. Booking Creation → Preauthorization
- Client creates booking
- Payment service creates preauthorization (holds funds)
- Client completes 3DS authentication
- Funds are held but not captured

### 2. Service Completion → Capture
- Cleaner marks booking as complete
- Admin/system approves completion
- Payment service captures preauthorized funds
- Money is transferred to platform account

### 3. Cancellation/Dispute → Refund
- If booking cancelled or dispute resolved in client favor
- Payment service processes refund
- Funds returned to client's card

---

## Netopia Payments Integration

### SDK Installation

```bash
go get github.com/netopiapayments/go-sdk
```

### Configuration

**Environment Variables** (`.env`):
```bash
# Netopia Sandbox
NETOPIA_POS_SIGNATURE=XXXX-XXXX-XXXX-XXXX-XXXX
NETOPIA_API_KEY=your-sandbox-api-key
NETOPIA_PUBLIC_KEY_PATH=./config/netopia-sandbox-public.key
NETOPIA_NOTIFY_URL=https://api-dev.cleanbuddy.ro/webhooks/netopia
NETOPIA_REDIRECT_URL=https://dev.cleanbuddy.ro/payment/result
ENV=development

# Netopia Production
# NETOPIA_POS_SIGNATURE=XXXX-XXXX-XXXX-XXXX-XXXX
# NETOPIA_API_KEY=your-production-api-key
# NETOPIA_PUBLIC_KEY_PATH=./config/netopia-production-public.key
# NETOPIA_NOTIFY_URL=https://api.cleanbuddy.ro/webhooks/netopia
# NETOPIA_REDIRECT_URL=https://cleanbuddy.ro/payment/result
# ENV=production
```

### Getting Credentials

1. **Sign up**: https://admin.netopia-payments.com/
2. **Sandbox Account**: Request sandbox credentials from support
3. **POS Signature**: Available in admin panel → Settings → API
4. **API Key**: Generate in admin panel → API Keys
5. **Public Key**: Download from admin panel → Security → Keys

### Code Implementation

**Initialize Netopia Client**:
```go
import (
    netopia "github.com/netopiapayments/go-sdk"
    "github.com/netopiapayments/go-sdk/requests"
)

func NewNetopiaClient() (*netopia.PaymentClient, error) {
    // Load public key
    publicKey, err := os.ReadFile(os.Getenv("NETOPIA_PUBLIC_KEY_PATH"))
    if err != nil {
        return nil, fmt.Errorf("failed to read public key: %w", err)
    }

    config := netopia.Config{
        PosSignature: os.Getenv("NETOPIA_POS_SIGNATURE"),
        ApiKey:       os.Getenv("NETOPIA_API_KEY"),
        IsLive:       os.Getenv("ENV") == "production",
        NotifyURL:    os.Getenv("NETOPIA_NOTIFY_URL"),
        RedirectURL:  os.Getenv("NETOPIA_REDIRECT_URL"),
        PublicKey:    publicKey,
    }

    // Create logger
    logger := &netopia.DefaultLogger{}

    client, err := netopia.NewPaymentClient(config, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create Netopia client: %w", err)
    }

    return client, nil
}
```

**Start Payment (Preauthorization)**:
```go
func (s *PaymentService) netopiaPreauthorize(payment *models.Payment) (*models.Payment, error) {
    client, err := NewNetopiaClient()
    if err != nil {
        return nil, err
    }

    // Get booking and client details
    booking, err := s.bookingRepo.GetByID(payment.BookingID)
    if err != nil {
        return nil, err
    }

    client, err := s.clientRepo.GetByUserID(payment.UserID)
    if err != nil {
        return nil, err
    }

    // Build payment request
    startReq := &requests.StartPaymentRequest{
        Payment: &requests.PaymentData{
            Options: requests.PaymentOptions{
                Installments: 1, // No installments
            },
            // Card details come from frontend form (not stored backend)
        },
        Order: &requests.OrderData{
            OrderID:     payment.ID,
            Amount:      payment.Amount,
            Currency:    "RON",
            Description: fmt.Sprintf("CleanBuddy - %s", booking.ServiceType),
            Billing: requests.BillingShipping{
                Email:     client.Email,
                FirstName: client.FirstName,
                LastName:  client.LastName,
            },
            Products: []requests.Product{
                {
                    Name:  fmt.Sprintf("Serviciu Curățenie - %s", booking.ServiceType),
                    Code:  string(booking.ServiceType),
                    Price: payment.Amount,
                    Vat:   19, // 19% VAT
                },
            },
        },
    }

    // Call Netopia API
    response, err := client.StartPayment(startReq)
    if err != nil {
        return nil, fmt.Errorf("Netopia StartPayment failed: %w", err)
    }

    // Update payment with response
    payment.ProviderTransactionID = sql.NullString{String: response.PaymentID, Valid: true}
    payment.ProviderOrderID = sql.NullString{String: startReq.Order.OrderID, Valid: true}
    payment.Status = models.PaymentStatusPending

    // Save payment
    err = s.paymentRepo.Create(payment)
    if err != nil {
        return nil, err
    }

    // Return payment with paymentURL for frontend redirect
    return payment, nil
}
```

**Check Payment Status**:
```go
func (s *PaymentService) netopiaGetStatus(paymentID string) (*models.Payment, error) {
    client, err := NewNetopiaClient()
    if err != nil {
        return nil, err
    }

    payment, err := s.paymentRepo.GetByID(paymentID)
    if err != nil {
        return nil, err
    }

    // Check status with Netopia
    ntpID := payment.ProviderTransactionID.String
    orderID := payment.ProviderOrderID.String

    statusResp, err := client.GetStatus(ntpID, orderID)
    if err != nil {
        return nil, err
    }

    // Update payment status based on Netopia response
    // statusResp contains payment state, error codes, etc.

    return payment, nil
}
```

**Handle IPN (Instant Payment Notification)**:
```go
func (s *PaymentService) HandleNetopiaIPN(r *http.Request) error {
    client, err := NewNetopiaClient()
    if err != nil {
        return err
    }

    // Verify IPN signature and extract data
    ipnResult, err := client.VerifyIPN(r)
    if err != nil {
        return fmt.Errorf("IPN verification failed: %w", err)
    }

    // Update payment in database based on IPN
    payment, err := s.paymentRepo.GetByProviderTransactionID(ipnResult.PaymentID)
    if err != nil {
        return err
    }

    switch ipnResult.ErrorCode {
    case "00": // Success
        payment.Status = models.PaymentStatusAuthorized
        payment.AuthorizedAt = sql.NullTime{Time: time.Now(), Valid: true}
    default: // Error
        payment.Status = models.PaymentStatusFailed
    }

    return s.paymentRepo.Update(payment)
}
```

### Testing

**Sandbox Test Cards**:
```
Successful Payment:
  Card: 9900009184214768
  Exp: 11/2025
  CVV: 111

3DS Authentication Required:
  Card: 9900004810289415
  Exp: 12/2025
  CVV: 222

Declined:
  Card: 9900002470636154
  Exp: 01/2026
  CVV: 333
```

**Test IPN**:
```bash
# Netopia will send IPN to your NETOPIA_NOTIFY_URL
# Test with sandbox transactions
# Verify signature with public key
```

---


## Webhook Endpoints

### Netopia IPN Handler

**Route**: `POST /webhooks/netopia`

```go
func (h *PaymentHandler) HandleNetopiaIPN(w http.ResponseWriter, r *http.Request) {
    err := h.paymentService.HandleNetopiaIPN(r)
    if err != nil {
        log.Printf("Netopia IPN error: %v", err)
        http.Error(w, "IPN processing failed", http.StatusInternalServerError)
        return
    }

    // Return 200 to acknowledge IPN
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```


## Frontend Integration

### Payment Form

```tsx
// frontend/src/components/PaymentForm.tsx
import { useState } from 'react';
import { useMutation, gql } from '@apollo/client';

const INITIALIZE_PAYMENT = gql`
  mutation InitializePayment($bookingId: ID!) {
    initializePayment(bookingId: $bookingId) {
      id
      status
      paymentURL
      providerTransactionID
    }
  }
`;

export function PaymentForm({ bookingId, amount }: Props) {
  const [initPayment, { loading }] = useMutation(INITIALIZE_PAYMENT);

  const handleSubmit = async () => {
    const { data } = await initPayment({
      variables: { bookingId },
    });

    // Redirect to Netopia for 3DS authentication
    if (data.initializePayment.paymentURL) {
      window.location.href = data.initializePayment.paymentURL;
    }
  };

  return (
    <div className="payment-form">
      <h2>Plată Sigură - {amount} RON</h2>
      <p className="text-gray-600">
        Plata este procesată securizat prin Netopia Payments
      </p>

      <button
        onClick={handleSubmit}
        disabled={loading}
        className="w-full bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700"
      >
        {loading ? 'Se procesează...' : 'Plătește Acum'}
      </button>

      <div className="security-badges flex justify-center gap-4 mt-4">
        <img src="/icons/visa.svg" alt="Visa" className="h-8" />
        <img src="/icons/mastercard.svg" alt="Mastercard" className="h-8" />
        <img src="/icons/3ds.svg" alt="3D Secure" className="h-8" />
      </div>

      <p className="text-xs text-gray-500 text-center mt-4">
        Plata ta este securizată cu criptare SSL și autentificare 3D Secure
      </p>
    </div>
  );
}
```

### Payment Result Page

```tsx
// frontend/src/app/payment/result/page.tsx
'use client';

import { useSearchParams } from 'next/navigation';
import { useEffect } from 'react';

export default function PaymentResultPage() {
  const searchParams = useSearchParams();
  const status = searchParams.get('status'); // success/failed
  const paymentId = searchParams.get('paymentId');

  useEffect(() => {
    // Verify payment status with backend
    // Update booking status
    // Show success/failure message
  }, [paymentId]);

  if (status === 'success') {
    return (
      <div className="success">
        <h1>✓ Plată Reușită!</h1>
        <p>Rezervarea ta a fost confirmată.</p>
        <a href="/dashboard/client">Vezi Rezervările</a>
      </div>
    );
  }

  return (
    <div className="error">
      <h1>✗ Plată Eșuată</h1>
      <p>A apărut o problemă cu plata ta.</p>
      <a href="/bookings">Încearcă Din Nou</a>
    </div>
  );
}
```

---

## Security Best Practices

### 1. Never Store Card Details
- Card numbers, CVV, exp dates MUST NOT be stored
- Use payment gateway's hosted forms
- CleanBuddy only stores: last 4 digits, card brand (for display)

### 2. Verify IPN Signatures
- Always verify Netopia signatures
- Use public key verification (Netopia)

- Reject unsigned/invalid IPNs

### 3. Idempotency
- Use payment.ID as idempotency key
- Prevent duplicate charges
- Handle IPN duplicates gracefully

### 4. HTTPS Only
- All payment endpoints must use HTTPS
- Webhook URLs must be HTTPS
- Redirect URLs must be HTTPS

### 5. PCI Compliance
- Use payment gateway SDKs/hosted forms
- Don't handle raw card data
- Log payment events (without card data)

---

## Testing Checklist

### Netopia Testing
- [ ] Successful payment with test card 9900009184214768
- [ ] 3DS authentication flow
- [ ] Declined payment with test card 9900002470636154
- [ ] IPN signature verification
- [ ] Payment status check
- [ ] Refund processing (manual or API)

### Integration Testing
- [ ] Full booking → payment → capture flow
- [ ] Cancellation → refund flow
- [ ] Dispute → refund flow
- [ ] Multiple concurrent payments
- [ ] Network failure handling
- [ ] IPN replay attack prevention

---

## Production Deployment

### Netopia
1. Get production credentials from Netopia support
2. Update .env with production values
3. Replace sandbox public key with production key
4. Update webhook URLs to production domain
5. Test with small real transaction
6. Monitor first 10 transactions closely

### Monitoring
- Set up alerts for payment failures
- Monitor IPN delivery rates
- Track payment success rates
- Log all payment events (Sentry)
- Daily reconciliation with gateway reports

---

## Troubleshooting

### Netopia Issues

**"Invalid signature"**
- Check public key path/content
- Verify POS signature matches
- Check API key is correct
- Ensure using correct environment (sandbox/live)

**"IPN not received"**
- Verify webhook URL is accessible (public HTTPS)
- Check firewall rules
- Test with ngrok for local development
- Check Netopia admin panel for IPN logs

**"Payment stuck in pending"**
- Check payment status with GetStatus API
- Verify 3DS completion
- Check IPN delivery
- Contact Netopia support with transaction ID

**"Invalid hash"**
- Check secret key
- Verify hash algorithm (HMAC-MD5)
- Ensure params in correct order
- Check for encoding issues

**"IPN validation failed"**
- Verify IPN params
- Check signature calculation
- Ensure XML response format correct

---

## Support Contacts

**Netopia**
- Support Email: support@netopia-payments.com
- Documentation: https://doc.netopia-payments.com/
- Admin Panel: https://admin.netopia-payments.com/

---

## Next Steps

1. **Complete Netopia Integration** (2-3 days)
   - Implement StartPayment with real SDK calls
   - Build IPN handler with signature verification
   - Test with sandbox environment
   - Add error handling and logging

   - Implement REST API calls
   - Build IPN handler
   - Test with sandbox
   - Add fallback logic

3. **Frontend Payment UI** (1-2 days)
   - Build payment form component
   - Add provider selection
   - Implement redirect flow
   - Create result pages

4. **Testing** (2-3 days)
   - End-to-end payment flows
   - Edge cases (network failures, timeouts)
   - Load testing
   - Security testing

5. **Production Deployment** (1 day)
   - Get production credentials
   - Deploy with monitoring
   - Test with real small transactions
   - Go live!

**Total Estimated Time**: 8-12 days for complete payment integration.
