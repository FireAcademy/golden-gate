package main

import (
    "time"
    "database/sql"
)

type User struct {
    RemainingCredits            int64          `json:"remaining_credits"`
    StripeCustomerID            sql.NullString `json:"stripe_customer_id"`
    AutoPurchaseCreditsPackages bool           `json:"auto_purchase_credits_packages"`
}

type APIKey struct {
    Disabled           bool   `json:"disabled"`
    MonthlyCreditLimit int64  `json:"monthly_credit_limit"`
    Origin             string `json:"origin"`
}

type Usage struct {
    Credits         int64     `json:"credits"`
}