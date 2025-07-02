#!/bin/bash

echo "Fixing ESLint errors in marketplace..."

# Fix APIDocumentation.tsx
sed -i '' "s/hasn't/hasn\&apos;t/g" src/components/APIDocumentation.tsx
sed -i '' 's/"Try it out"/\&quot;Try it out\&quot;/g' src/components/APIDocumentation.tsx

# Fix forgot-password.tsx
sed -i '' "s/doesn't/doesn\&apos;t/g" src/pages/auth/forgot-password.tsx

# Fix verify-email.tsx
sed -i '' "s/didn't/didn\&apos;t/g" src/pages/auth/verify-email.tsx
sed -i '' "s/We've/We\&apos;ve/g" src/pages/auth/verify-email.tsx

# Fix authentication.tsx
sed -i '' 's/"api_key"/\&quot;api_key\&quot;/g' src/pages/docs/authentication.tsx
sed -i '' 's/"YOUR_API_KEY"/\&quot;YOUR_API_KEY\&quot;/g' src/pages/docs/authentication.tsx
sed -i '' 's/"X-API-Key"/\&quot;X-API-Key\&quot;/g' src/pages/docs/authentication.tsx
sed -i '' "s/doesn't/doesn\&apos;t/g" src/pages/docs/authentication.tsx
sed -i '' "s/you've/you\&apos;ve/g" src/pages/docs/authentication.tsx
sed -i '' "s/they're/they\&apos;re/g" src/pages/docs/authentication.tsx
sed -i '' "s/we'll/we\&apos;ll/g" src/pages/docs/authentication.tsx

# Fix getting-started.tsx
sed -i '' 's/"api_key"/\&quot;api_key\&quot;/g' src/pages/docs/getting-started.tsx
sed -i '' 's/"YOUR_API_KEY"/\&quot;YOUR_API_KEY\&quot;/g' src/pages/docs/getting-started.tsx
sed -i '' "s/You'll/You\&apos;ll/g" src/pages/docs/getting-started.tsx
sed -i '' "s/you're/you\&apos;re/g" src/pages/docs/getting-started.tsx
sed -i '' "s/Here's/Here\&apos;s/g" src/pages/docs/getting-started.tsx

# Fix sdks.tsx
sed -i '' "s/don't/don\&apos;t/g" src/pages/docs/sdks.tsx
sed -i '' "s/we're/we\&apos;re/g" src/pages/docs/sdks.tsx

# Fix support.tsx
sed -i '' "s/we'll/we\&apos;ll/g" src/pages/docs/support.tsx

echo "ESLint fixes applied. Running lint check..."
npm run lint