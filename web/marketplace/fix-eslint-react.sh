#!/bin/bash

echo "Fixing ESLint errors with proper React escaping..."

# Fix forgot-password.tsx
sed -i '' "s/doesn't/doesn\{'\}t/g" src/pages/auth/forgot-password.tsx

# Fix verify-email.tsx  
sed -i '' "s/Didn't/Didn\{'\}t/g" src/pages/auth/verify-email.tsx

# Fix authentication.tsx
sed -i '' 's/"api_key"/{\\"api_key\\"}/g' src/pages/docs/authentication.tsx
sed -i '' 's/"grant_type"/{\\"grant_type\\"}/g' src/pages/docs/authentication.tsx
sed -i '' "s/doesn't/doesn\{'\}t/g" src/pages/docs/authentication.tsx
sed -i '' "s/you've/you\{'\}ve/g" src/pages/docs/authentication.tsx
sed -i '' "s/they're/they\{'\}re/g" src/pages/docs/authentication.tsx
sed -i '' "s/we'll/we\{'\}ll/g" src/pages/docs/authentication.tsx

# Fix getting-started.tsx
sed -i '' 's/"api_key"/{\\"api_key\\"}/g' src/pages/docs/getting-started.tsx
sed -i '' 's/"YOUR_API_KEY"/{\\"YOUR_API_KEY\\"}/g' src/pages/docs/getting-started.tsx
sed -i '' "s/You'll/You\{'\}ll/g" src/pages/docs/getting-started.tsx
sed -i '' "s/you're/you\{'\}re/g" src/pages/docs/getting-started.tsx
sed -i '' "s/Here's/Here\{'\}s/g" src/pages/docs/getting-started.tsx

# Fix sdks.tsx
sed -i '' "s/we're/we\{'\}re/g" src/pages/docs/sdks.tsx

# Fix support.tsx
sed -i '' "s/we'll/we\{'\}ll/g" src/pages/docs/support.tsx

# Fix index.tsx
sed -i '' "s/Let's/Let\{'\}s/g" src/pages/index.tsx
sed -i '' "s/You'll/You\{'\}ll/g" src/pages/index.tsx

# Fix subscribe/[apiId].tsx
sed -i '' "s/you'll/you\{'\}ll/g" src/pages/subscribe/\[apiId\].tsx

echo "ESLint fixes applied. Running lint check..."
npm run lint