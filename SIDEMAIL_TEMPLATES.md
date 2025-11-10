# Sidemail Email Templates Documentation

This document lists all email templates that need to be created in Sidemail for the CleanBuddy platform.

## Setup Instructions

1. Log in to [Sidemail Dashboard](https://app.sidemail.io)
2. Navigate to Templates section
3. Create each template below with the specified template name and variables

## Required Templates

### 1. booking-confirmation

**Template Name**: `booking-confirmation`

**Description**: Sent to client when they create a new booking

**Template Variables**:
- `clientName` (string) - Client's first name
- `bookingID` (string) - Booking ID
- `serviceType` (string) - Type of cleaning service (e.g., "STANDARD", "DEEP_CLEAN")
- `scheduledDate` (string) - Date in format "02 January 2006"
- `scheduledTime` (string) - Time in format "15:04"
- `totalPrice` (string) - Price with currency (e.g., "150.00 RON")

**Email Subject**: `Confirmare rezervare #{{bookingID}}`

**Sample Content**:
```
Bună {{clientName}},

Rezervarea ta a fost înregistrată cu succes!

Detalii rezervare:
- ID Rezervare: #{{bookingID}}
- Tip serviciu: {{serviceType}}
- Dată programată: {{scheduledDate}}
- Oră: {{scheduledTime}}
- Preț total: {{totalPrice}}

Vei primi o notificare când un partener de curățenie va accepta rezervarea.

Cu respect,
Echipa CleanBuddy
```

---

### 2. booking-accepted

**Template Name**: `booking-accepted`

**Description**: Sent to client when a cleaner accepts their booking

**Template Variables**:
- `clientName` (string) - Client's first name
- `cleanerName` (string) - Cleaner's first name
- `bookingID` (string) - Booking ID
- `scheduledDate` (string) - Date in format "02 January 2006"
- `scheduledTime` (string) - Time in format "15:04"

**Email Subject**: `Rezervarea ta #{{bookingID}} a fost acceptată`

**Sample Content**:
```
Bună {{clientName}},

Vești bune! Rezervarea ta a fost acceptată de {{cleanerName}}.

Detalii:
- ID Rezervare: #{{bookingID}}
- Data: {{scheduledDate}}
- Ora: {{scheduledTime}}
- Partener: {{cleanerName}}

{{cleanerName}} va ajunge la adresa indicată la ora programată.

Cu respect,
Echipa CleanBuddy
```

---

### 3. booking-cancelled

**Template Name**: `booking-cancelled`

**Description**: Sent when a booking is cancelled by client or cleaner

**Template Variables**:
- `recipientName` (string) - Name of email recipient
- `bookingID` (string) - Booking ID
- `cancelledBy` (string) - Who cancelled: "You", "the client", or "the cleaner"
- `cancellationReason` (string) - Reason for cancellation

**Email Subject**: `Rezervare anulată #{{bookingID}}`

**Sample Content**:
```
Bună {{recipientName}},

Rezervarea #{{bookingID}} a fost anulată de {{cancelledBy}}.

Motiv: {{cancellationReason}}

Dacă ai întrebări, nu ezita să ne contactezi.

Cu respect,
Echipa CleanBuddy
```

---

### 4. booking-completed

**Template Name**: `booking-completed`

**Description**: Sent to client when booking is marked as completed

**Template Variables**:
- `clientName` (string) - Client's first name
- `bookingID` (string) - Booking ID
- `totalPrice` (string) - Final price with currency
- `reviewURL` (string) - URL to review page

**Email Subject**: `Serviciul tău a fost finalizat #{{bookingID}}`

**Sample Content**:
```
Bună {{clientName}},

Serviciul de curățenie pentru rezervarea #{{bookingID}} a fost finalizat cu succes!

Suma totală: {{totalPrice}}

Ne-ar plăcea să aflăm părerea ta despre serviciu:
{{reviewURL}}

Mulțumim că folosești CleanBuddy!

Cu respect,
Echipa CleanBuddy
```

---

### 5. cleaner-approved

**Template Name**: `cleaner-approved`

**Description**: Sent when admin approves a cleaner's profile

**Template Variables**:
- `cleanerName` (string) - Cleaner's first name

**Email Subject**: `Profilul tău a fost aprobat!`

**Sample Content**:
```
Felicitări {{cleanerName}},

Profilul tău de partener de curățenie a fost aprobat!

Poți acum să primești și să accepți rezervări de la clienți.

Mergi la dashboard-ul tău pentru a vedea job-uri disponibile și a-ți configura disponibilitatea.

Mult succes!

Echipa CleanBuddy
```

---

### 6. cleaner-rejected

**Template Name**: `cleaner-rejected`

**Description**: Sent when admin rejects a cleaner's profile

**Template Variables**:
- `cleanerName` (string) - Cleaner's first name
- `rejectionReason` (string) - Reason for rejection

**Email Subject**: `Actualizare aplicație CleanBuddy`

**Sample Content**:
```
Bună {{cleanerName}},

Din păcate, aplicația ta de partener de curățenie nu a fost aprobată în acest moment.

Motiv: {{rejectionReason}}

Dacă dorești să corectezi informațiile și să reaplici, te rugăm să ne contactezi la support@cleanbuddy.ro.

Cu respect,
Echipa CleanBuddy
```

---

### 7. payout-processed

**Template Name**: `payout-processed`

**Description**: Sent to cleaner when monthly payout is processed

**Template Variables**:
- `cleanerName` (string) - Cleaner's first name
- `amount` (string) - Payout amount with currency
- `period` (string) - Period (e.g., "01 January 2025 - 31 January 2025")
- `transferRef` (string) - Bank transfer reference number

**Email Subject**: `Plata ta lunară a fost procesată`

**Sample Content**:
```
Bună {{cleanerName}},

Plata ta pentru perioada {{period}} a fost procesată.

Suma: {{amount}}
Referință transfer: {{transferRef}}

Banii vor ajunge în contul tău în 1-3 zile lucrătoare.

Mulțumim pentru munca ta!

Echipa CleanBuddy
```

---

### 8. welcome-email

**Template Name**: `welcome-email`

**Description**: Sent when a new user signs up (optional, for future use)

**Template Variables**:
- `userName` (string) - User's first name
- `userRole` (string) - User role ("CLIENT" or "CLEANER")

**Email Subject**: `Bun venit la CleanBuddy!`

**Sample Content**:
```
Bună {{userName}},

Bine ai venit la CleanBuddy - platforma ta de servicii profesionale de curățenie!

{{#if userRole == "CLIENT"}}
Poți acum să rezervi servicii de curățenie de la parteneri verificați.
{{else}}
Poți acum să primești job-uri de curățenie și să îți gestionezi programul.
{{/if}}

Explorează dashboard-ul tău pentru a începe.

Cu respect,
Echipa CleanBuddy
```

---

## Testing Templates

After creating all templates in Sidemail:

1. Update `.env` file with your Sidemail API key:
   ```bash
   SIDEMAIL_API_KEY=your-actual-api-key
   EMAIL_FROM_ADDRESS=noreply@cleanbuddy.ro
   EMAIL_FROM_NAME=CleanBuddy
   ```

2. Test email sending using GraphQL mutations or direct API calls

3. Check Sidemail dashboard for delivery status

## Notes

- All templates should use Romanian language
- Use professional, friendly tone
- Include CleanBuddy branding (logo, colors)
- Add footer with contact info and unsubscribe link
- Test with sample data before going live
- Monitor delivery rates and open rates in Sidemail dashboard
