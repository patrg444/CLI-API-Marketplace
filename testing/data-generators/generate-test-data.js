const { faker } = require('@faker-js/faker');
const fs = require('fs').promises;
const path = require('path');

// Configuration
const CONFIG = {
  numCreators: 50,
  numConsumers: 200,
  numAPIs: 100,
  numReviews: 500,
  outputDir: './test-data'
};

// Categories and tags
const CATEGORIES = [
  'Financial Services',
  'AI/ML',
  'Communication',
  'Data & Analytics',
  'Developer Tools',
  'Security',
  'Media',
  'IoT',
  'Healthcare',
  'Education'
];

const TAGS = {
  'Financial Services': ['payment', 'stripe', 'billing', 'invoice', 'accounting', 'crypto', 'banking'],
  'AI/ML': ['machine-learning', 'nlp', 'computer-vision', 'tensorflow', 'pytorch', 'prediction'],
  'Communication': ['email', 'sms', 'chat', 'video', 'voice', 'notification', 'messaging'],
  'Data & Analytics': ['analytics', 'reporting', 'visualization', 'bigdata', 'etl', 'metrics'],
  'Developer Tools': ['ci-cd', 'testing', 'monitoring', 'logging', 'debugging', 'deployment'],
  'Security': ['authentication', 'encryption', 'auth0', 'oauth', 'jwt', 'security-scanning'],
  'Media': ['image', 'video', 'audio', 'streaming', 'transcoding', 'cdn'],
  'IoT': ['sensors', 'telemetry', 'mqtt', 'device-management', 'edge-computing'],
  'Healthcare': ['ehr', 'fhir', 'medical-imaging', 'telemedicine', 'health-data'],
  'Education': ['lms', 'assessment', 'content', 'student-tracking', 'e-learning']
};

// Pricing tiers
const PRICING_PLANS = [
  {
    name: 'Free',
    type: 'free',
    monthly_price: 0,
    call_limit: 1000,
    rate_limit_per_minute: 10
  },
  {
    name: 'Basic',
    type: 'subscription',
    monthly_price: 29.99,
    call_limit: 50000,
    rate_limit_per_minute: 60
  },
  {
    name: 'Pro',
    type: 'subscription',
    monthly_price: 99.99,
    call_limit: 500000,
    rate_limit_per_minute: 100
  },
  {
    name: 'Enterprise',
    type: 'subscription',
    monthly_price: 499.99,
    call_limit: -1, // unlimited
    rate_limit_per_minute: 1000
  }
];

// Generate creators
function generateCreators(count) {
  const creators = [];
  for (let i = 0; i < count; i++) {
    creators.push({
      id: `creator_${i + 1}`,
      email: faker.internet.email(),
      name: faker.company.name(),
      bio: faker.company.catchPhrase(),
      website: faker.internet.url(),
      stripe_account_id: `acct_${faker.string.alphanumeric(16)}`,
      created_at: faker.date.past(2),
      verified: Math.random() > 0.2
    });
  }
  return creators;
}

// Generate consumers
function generateConsumers(count) {
  const consumers = [];
  for (let i = 0; i < count; i++) {
    consumers.push({
      id: `consumer_${i + 1}`,
      email: faker.internet.email(),
      name: faker.person.fullName(),
      company: Math.random() > 0.5 ? faker.company.name() : null,
      stripe_customer_id: `cus_${faker.string.alphanumeric(14)}`,
      created_at: faker.date.past(2),
      email_verified: Math.random() > 0.1
    });
  }
  return consumers;
}

// Generate APIs
function generateAPIs(count, creators) {
  const apis = [];
  for (let i = 0; i < count; i++) {
    const category = faker.helpers.arrayElement(CATEGORIES);
    const tags = faker.helpers.arrayElements(TAGS[category], { min: 2, max: 5 });
    const creator = faker.helpers.arrayElement(creators);
    const hasFreeTier = Math.random() > 0.3;
    
    // Select pricing plans
    let plans = hasFreeTier 
      ? [PRICING_PLANS[0], ...faker.helpers.arrayElements(PRICING_PLANS.slice(1), { min: 1, max: 3 })]
      : faker.helpers.arrayElements(PRICING_PLANS.slice(1), { min: 1, max: 3 });
    
    apis.push({
      id: `api_${i + 1}`,
      name: faker.helpers.fake('{{company.buzzAdjective}} {{company.buzzNoun}} API'),
      slug: faker.helpers.slugify(faker.company.buzzPhrase()).toLowerCase(),
      description: faker.lorem.paragraph(),
      long_description: faker.lorem.paragraphs(3),
      category: category,
      tags: tags,
      creator_id: creator.id,
      base_url: faker.internet.url(),
      documentation_url: faker.internet.url() + '/docs',
      version: faker.system.semver(),
      status: 'published',
      pricing_plans: plans,
      features: [
        faker.company.buzzPhrase(),
        faker.company.buzzPhrase(),
        faker.company.buzzPhrase()
      ],
      created_at: faker.date.past(1),
      published_at: faker.date.recent(90),
      total_subscriptions: faker.number.int({ min: 0, max: 1000 }),
      rating: faker.number.float({ min: 3.0, max: 5.0, precision: 0.1 }),
      total_reviews: faker.number.int({ min: 0, max: 100 })
    });
  }
  return apis;
}

// Generate reviews
function generateReviews(count, apis, consumers) {
  const reviews = [];
  const reviewedPairs = new Set(); // To prevent duplicate reviews
  
  for (let i = 0; i < count; i++) {
    const api = faker.helpers.arrayElement(apis);
    const consumer = faker.helpers.arrayElement(consumers);
    const pairKey = `${api.id}-${consumer.id}`;
    
    // Skip if this consumer already reviewed this API
    if (reviewedPairs.has(pairKey)) {
      continue;
    }
    reviewedPairs.add(pairKey);
    
    const rating = faker.number.int({ min: 1, max: 5 });
    const hasCreatorResponse = Math.random() > 0.7;
    
    reviews.push({
      id: `review_${i + 1}`,
      api_id: api.id,
      consumer_id: consumer.id,
      rating: rating,
      title: generateReviewTitle(rating),
      comment: generateReviewComment(rating),
      verified_purchase: Math.random() > 0.2,
      helpful_count: faker.number.int({ min: 0, max: 50 }),
      not_helpful_count: faker.number.int({ min: 0, max: 10 }),
      created_at: faker.date.recent(60),
      creator_response: hasCreatorResponse ? {
        comment: faker.lorem.paragraph(),
        created_at: faker.date.recent(30)
      } : null
    });
  }
  return reviews;
}

// Generate review titles based on rating
function generateReviewTitle(rating) {
  const titles = {
    5: [
      'Excellent API!',
      'Highly Recommended',
      'Perfect for our needs',
      'Outstanding service',
      'Best in class'
    ],
    4: [
      'Great API with minor issues',
      'Very good overall',
      'Solid choice',
      'Works well',
      'Good value'
    ],
    3: [
      'Decent but could improve',
      'Average experience',
      'Gets the job done',
      'Mixed feelings',
      'OK for basic needs'
    ],
    2: [
      'Disappointing',
      'Needs improvement',
      'Below expectations',
      'Several issues',
      'Not recommended'
    ],
    1: [
      'Poor experience',
      'Avoid',
      'Many problems',
      'Terrible',
      'Complete waste'
    ]
  };
  
  return faker.helpers.arrayElement(titles[rating]);
}

// Generate review comments based on rating
function generateReviewComment(rating) {
  if (rating >= 4) {
    return faker.helpers.fake(
      'The {{company.buzzAdjective}} features are excellent. ' +
      '{{lorem.sentence}} ' +
      'The documentation is {{company.buzzAdjective}} and the support team is very responsive. ' +
      '{{lorem.sentence}}'
    );
  } else if (rating === 3) {
    return faker.helpers.fake(
      'The API works but has some limitations. ' +
      '{{lorem.sentence}} ' +
      'Documentation could be better and {{lorem.sentence}} ' +
      'It\'s okay for basic use cases.'
    );
  } else {
    return faker.helpers.fake(
      'Unfortunately, this API has several issues. ' +
      '{{lorem.sentence}} ' +
      'The {{company.buzzNoun}} feature doesn\'t work as advertised. ' +
      '{{lorem.sentence}} Not worth the price.'
    );
  }
}

// Generate API keys for consumers
function generateAPIKeys(apis, consumers) {
  const apiKeys = [];
  const subscriptions = [];
  
  // Each consumer subscribes to 0-5 APIs
  consumers.forEach(consumer => {
    const numSubscriptions = faker.number.int({ min: 0, max: 5 });
    const subscribedAPIs = faker.helpers.arrayElements(apis, numSubscriptions);
    
    subscribedAPIs.forEach(api => {
      const plan = faker.helpers.arrayElement(api.pricing_plans);
      const subscriptionId = `sub_${faker.string.alphanumeric(14)}`;
      
      subscriptions.push({
        id: subscriptionId,
        consumer_id: consumer.id,
        api_id: api.id,
        plan_name: plan.name,
        status: 'active',
        stripe_subscription_id: plan.type !== 'free' ? `sub_${faker.string.alphanumeric(14)}` : null,
        created_at: faker.date.recent(180),
        current_period_start: faker.date.recent(30),
        current_period_end: faker.date.future(1)
      });
      
      apiKeys.push({
        id: `key_${faker.string.alphanumeric(32)}`,
        consumer_id: consumer.id,
        api_id: api.id,
        subscription_id: subscriptionId,
        key_hash: faker.string.alphanumeric(64),
        name: `${api.name} - ${plan.name}`,
        created_at: faker.date.recent(180),
        last_used_at: faker.date.recent(1),
        total_calls: faker.number.int({ min: 0, max: plan.call_limit === -1 ? 100000 : (plan.call_limit || 100000) })
      });
    });
  });
  
  return { apiKeys, subscriptions };
}

// Generate usage data
function generateUsageData(apiKeys) {
  const usage = [];
  
  apiKeys.forEach(key => {
    // Generate usage for last 30 days
    for (let i = 0; i < 30; i++) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      
      usage.push({
        api_key_id: key.id,
        date: date.toISOString().split('T')[0],
        calls: faker.number.int({ min: 0, max: 1000 }),
        errors: faker.number.int({ min: 0, max: 50 }),
        latency_p50: faker.number.int({ min: 20, max: 100 }),
        latency_p95: faker.number.int({ min: 50, max: 500 }),
        latency_p99: faker.number.int({ min: 100, max: 1000 })
      });
    }
  });
  
  return usage;
}

// Save data to files
async function saveData(data, filename) {
  const outputPath = path.join(CONFIG.outputDir, filename);
  await fs.writeFile(outputPath, JSON.stringify(data, null, 2));
  console.log(`Saved ${data.length} records to ${outputPath}`);
}

// Main function
async function generateTestData() {
  console.log('🚀 Generating test data...\n');
  
  // Create output directory
  await fs.mkdir(CONFIG.outputDir, { recursive: true });
  
  // Generate data
  console.log('Creating creators...');
  const creators = generateCreators(CONFIG.numCreators);
  await saveData(creators, 'creators.json');
  
  console.log('Creating consumers...');
  const consumers = generateConsumers(CONFIG.numConsumers);
  await saveData(consumers, 'consumers.json');
  
  console.log('Creating APIs...');
  const apis = generateAPIs(CONFIG.numAPIs, creators);
  await saveData(apis, 'apis.json');
  
  console.log('Creating reviews...');
  const reviews = generateReviews(CONFIG.numReviews, apis, consumers);
  await saveData(reviews, 'reviews.json');
  
  console.log('Creating subscriptions and API keys...');
  const { apiKeys, subscriptions } = generateAPIKeys(apis, consumers);
  await saveData(subscriptions, 'subscriptions.json');
  await saveData(apiKeys, 'api_keys.json');
  
  console.log('Creating usage data...');
  const usage = generateUsageData(apiKeys.slice(0, 100)); // Limit usage data
  await saveData(usage, 'usage.json');
  
  // Generate summary
  const summary = {
    generated_at: new Date().toISOString(),
    counts: {
      creators: creators.length,
      consumers: consumers.length,
      apis: apis.length,
      reviews: reviews.length,
      subscriptions: subscriptions.length,
      api_keys: apiKeys.length,
      usage_records: usage.length
    },
    categories: CATEGORIES,
    sample_api_names: apis.slice(0, 10).map(api => api.name)
  };
  
  await saveData([summary], 'summary.json');
  
  console.log('\n✅ Test data generation complete!');
  console.log(`\nData saved to: ${path.resolve(CONFIG.outputDir)}`);
  console.log('\nSummary:');
  console.log(JSON.stringify(summary.counts, null, 2));
}

// Run if called directly
if (require.main === module) {
  generateTestData().catch(console.error);
}

module.exports = {
  generateCreators,
  generateConsumers,
  generateAPIs,
  generateReviews,
  generateAPIKeys,
  generateUsageData
};
