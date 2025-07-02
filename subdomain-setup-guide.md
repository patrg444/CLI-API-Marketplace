# Complete Subdomain Setup Guide for apidirect.dev

## Step 1: Add DNS Records in Namecheap

1. **Login to Namecheap**
   - Go to namecheap.com and login
   - Click on "Domain List"
   - Find `apidirect.dev` and click "Manage"

2. **Go to Advanced DNS**
   - Click on "Advanced DNS" tab

3. **Add A Records for Each Subdomain**
   
   Click "Add New Record" and add these:
   
   | Type | Host | Value | TTL |
   |------|------|-------|-----|
   | A Record | console | 216.198.79.193 | Automatic |
   | A Record | api | 216.198.79.193 | Automatic |
   | A Record | docs | 216.198.79.193 | Automatic |

   **Note**: 216.198.79.193 is Vercel's IP address

## Step 2: Configure Subdomains in Vercel

### For console.apidirect.dev:

1. **Go to your Vercel Dashboard**
   - Visit vercel.com/dashboard
   - Find your `cli-marketplace` project

2. **Add Domain**
   - Go to Settings → Domains
   - Click "Add Domain"
   - Enter: `console.apidirect.dev`
   - Click "Add"

3. **Configure Root Directory**
   - Go to Settings → General
   - Find "Root Directory"
   - Change it to: `web/console`
   - Save changes

### Alternative: Deploy as Separate Projects (Recommended)

This is cleaner for managing different parts:

```bash
# Deploy console as separate project
cd /Users/patrickgloria/CLI-API-Marketplace/web/console
vercel

# When prompted:
# - Set up and deploy: Y
# - Which scope: (your account)
# - Link to existing project: N
# - Project name: apidirect-console
# - Root directory: ./
# - Override settings: N
```

Then add the domain in Vercel dashboard.

## Step 3: Wait for DNS Propagation

- Takes 5-30 minutes typically
- Maximum 48 hours (rare)

## Step 4: Test Your Subdomains

```bash
# Test DNS resolution
nslookup console.apidirect.dev
nslookup api.apidirect.dev

# Test with curl
curl -I https://console.apidirect.dev
```

## Deployment Structure

### Option A: Single Vercel Project with Rewrites
```
vercel.json (in root):
{
  "rewrites": [
    {
      "source": "/",
      "destination": "/web/landing/index.html"
    },
    {
      "source": "/console/(.*)",
      "destination": "/web/console/$1"
    }
  ]
}
```

### Option B: Separate Vercel Projects (Recommended)
- `apidirect-landing` → apidirect.dev
- `apidirect-console` → console.apidirect.dev
- `apidirect-api` → api.apidirect.dev

## Quick Commands for Separate Deployments

```bash
# Deploy landing page
cd web/landing
vercel --prod

# Deploy console
cd ../console
vercel --prod

# Deploy API (when ready)
cd ../../backend
vercel --prod
```

## Troubleshooting

1. **"Invalid Configuration" in Vercel**
   - Make sure DNS records point to 216.198.79.193
   - Wait for DNS propagation

2. **SSL Certificate Issues**
   - Vercel automatically provisions SSL
   - May take 10-15 minutes after domain verification

3. **404 Errors**
   - Check root directory setting
   - Ensure files are pushed to GitHub

## Next Steps

1. ✅ Add DNS records in Namecheap (5 min)
2. ✅ Deploy console to Vercel (10 min)
3. ✅ Add domain in Vercel settings (5 min)
4. ✅ Wait for propagation (5-30 min)
5. ✅ Access console.apidirect.dev!

Need help? The Vercel dashboard shows domain status and any issues.