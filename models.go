package main

import (
    "database/sql"
)

type UserInfo struct {
    RemainingCredits            int64          `json:"remaining_credits"`
    StripeCustomerID            sql.NullString `json:"stripe_customer_id"`
    AutoPurchaseCreditsPackages bool           `json:"auto_purchase_credits_packages"`
}

type APIKeyInfo struct {
    Disabled           bool   `json:"disabled"`
    MonthlyCreditLimit int64  `json:"monthly_credit_limit"`
    Origin             string `json:"origin"`
}

type UsageInfo struct {
    Credits int64 `json:"credits"`
}

type DataDudeResponse struct {
    Success    bool       `json:"success"`
    APIKey     APIKeyInfo `json:"api_key"`
    User       UserInfo   `json:"user"`
    Usage      UsageInfo  `json:"usage"`
}