# Subdomain Setup Guide for API-Direct

## Current Domain Structure

You already own `apidirect.dev`. Here's how to set up subdomains:

### 1. In Namecheap (your registrar):
Add these DNS records:
```
Type    Name        Value
A       console     76.76.21.21 (Vercel IP)
A       api         76.76.21.21 (Vercel IP)
A       docs        76.76.21.21 (Vercel IP)
```

### 2. In Vercel:
Add each subdomain to your project:
- console.apidirect.dev
- api.apidirect.dev
- docs.apidirect.dev

### 3. Deploy Structure:

```
/CLI-API-Marketplace
├── web/
│   ├── landing/        → apidirect.dev
│   ├── console/        → console.apidirect.dev
│   └── docs/           → docs.apidirect.dev (optional)
├── backend/            → api.apidirect.dev
└── services/
```

## Benefits of Using Subdomains:

1. **Professional appearance** - Everything under one brand
2. **Cost effective** - No additional domain purchases
3. **Better SEO** - Search engines see it as one property
4. **Easier SSL** - Wildcard certificate covers all subdomains
5. **User trust** - Consistent domain builds confidence

## Current References in Your Code:

Your landing page already references:
- `https://console.api-direct.io` → Change to `console.apidirect.dev`
- `https://api.yourdomain.com` → Change to `api.apidirect.dev`

## No Additional Domains Needed! ✅

Everything runs under your single `apidirect.dev` domain using subdomains.