import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControl,
  FormControlLabel,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import {
  Save as SaveIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Upload as UploadIcon,
  AttachMoney as MoneyIcon,
  Description as DocsIcon,
  Visibility as PreviewIcon,
} from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';

const API_CATEGORIES = [
  'AI/ML',
  'Data',
  'Communication',
  'Finance',
  'Media',
  'Security',
  'Analytics',
  'Productivity',
  'Developer Tools',
  'Other'
];

const PRICING_TYPES = {
  free: 'Free',
  pay_per_use: 'Pay per use',
  subscription: 'Monthly subscription'
};

function MarketplaceSettings() {
  const { apiId } = useParams();
  const navigate = useNavigate();
  const [tabValue, setTabValue] = useState(0);
  const [loading, setLoading] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // General settings state
  const [isPublished, setIsPublished] = useState(false);
  const [apiName, setApiName] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [tags, setTags] = useState([]);
  const [newTag, setNewTag] = useState('');

  // Pricing state
  const [pricingPlans, setPricingPlans] = useState([
    {
      id: 'plan-1',
      name: 'Free Tier',
      type: 'free',
      price_per_call: 0,
      monthly_price: 0,
      call_limit: 1000,
      rate_limit_per_minute: 10,
      rate_limit_per_day: 1000,
      is_active: true
    }
  ]);
  const [planDialog, setPlanDialog] = useState({ open: false, plan: null, isNew: false });

  // Documentation state
  const [openApiSpec, setOpenApiSpec] = useState(null);
  const [markdownDocs, setMarkdownDocs] = useState('');

  useEffect(() => {
    fetchAPIDetails();
  }, [apiId]);

  const fetchAPIDetails = async () => {
    setLoading(true);
    try {
      // Mock data for now
      setTimeout(() => {
        setApiName('Weather API');
        setIsPublished(true);
        setDescription('A comprehensive weather API providing real-time weather data, forecasts, and historical weather information for locations worldwide.');
        setCategory('Data');
        setTags(['weather', 'forecast', 'climate', 'real-time']);
        setLoading(false);
      }, 500);
    } catch (error) {
      console.error('Failed to fetch API details:', error);
      setLoading(false);
    }
  };

  const handleSaveGeneral = async () => {
    setLoading(true);
    try {
      // Save general settings
      await new Promise(resolve => setTimeout(resolve, 1000));
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddTag = () => {
    if (newTag && !tags.includes(newTag)) {
      setTags([...tags, newTag]);
      setNewTag('');
    }
  };

  const handleRemoveTag = (tagToRemove) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleSavePlan = () => {
    const { plan, isNew } = planDialog;
    if (isNew) {
      setPricingPlans([...pricingPlans, { ...plan, id: `plan-${Date.now()}` }]);
    } else {
      setPricingPlans(pricingPlans.map(p => p.id === plan.id ? plan : p));
    }
    setPlanDialog({ open: false, plan: null, isNew: false });
  };

  const handleDeletePlan = (planId) => {
    setPricingPlans(pricingPlans.filter(p => p.id !== planId));
  };

  const handleFileUpload = (event, type) => {
    const file = event.target.files[0];
    if (file) {
      if (type === 'openapi') {
        // Handle OpenAPI spec upload
        const reader = new FileReader();
        reader.onload = (e) => {
          try {
            const spec = JSON.parse(e.target.result);
            setOpenApiSpec(spec);
          } catch (error) {
            console.error('Invalid OpenAPI spec:', error);
          }
        };
        reader.readAsText(file);
      }
    }
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography variant="h4" fontWeight="bold">
            Marketplace Settings
          </Typography>
          <Typography variant="body2" color="text.secondary" mt={1}>
            {apiName} - Configure how your API appears in the marketplace
          </Typography>
        </Box>
        <Box display="flex" gap={2}>
          <Button
            variant="outlined"
            startIcon={<PreviewIcon />}
            onClick={() => window.open(`/marketplace/api/${apiId}`, '_blank')}
          >
            Preview Listing
          </Button>
          <FormControlLabel
            control={
              <Switch
                checked={isPublished}
                onChange={(e) => setIsPublished(e.target.checked)}
                color="primary"
              />
            }
            label={isPublished ? 'Published' : 'Not Published'}
          />
        </Box>
      </Box>

      {saveSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Settings saved successfully!
        </Alert>
      )}

      <Paper>
        <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)}>
          <Tab label="General" />
          <Tab label="Pricing Plans" icon={<MoneyIcon />} />
          <Tab label="Documentation" icon={<DocsIcon />} />
        </Tabs>

        {/* General Tab */}
        {tabValue === 0 && (
          <Box p={3}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="API Description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  multiline
                  rows={4}
                  helperText="A detailed description of what your API does and its key features"
                />
              </Grid>

              <Grid item xs={12} md={6}>
                <FormControl fullWidth>
                  <InputLabel>Category</InputLabel>
                  <Select
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    label="Category"
                  >
                    {API_CATEGORIES.map((cat) => (
                      <MenuItem key={cat} value={cat}>{cat}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>

              <Grid item xs={12}>
                <Typography variant="subtitle2" gutterBottom>
                  Tags
                </Typography>
                <Box display="flex" flexWrap="wrap" gap={1} mb={2}>
                  {tags.map((tag) => (
                    <Chip
                      key={tag}
                      label={tag}
                      onDelete={() => handleRemoveTag(tag)}
                      color="primary"
                      variant="outlined"
                    />
                  ))}
                </Box>
                <Box display="flex" gap={1}>
                  <TextField
                    size="small"
                    placeholder="Add a tag"
                    value={newTag}
                    onChange={(e) => setNewTag(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleAddTag()}
                  />
                  <Button
                    variant="outlined"
                    size="small"
                    onClick={handleAddTag}
                    startIcon={<AddIcon />}
                  >
                    Add
                  </Button>
                </Box>
              </Grid>

              <Grid item xs={12}>
                <Box display="flex" justifyContent="flex-end">
                  <Button
                    variant="contained"
                    startIcon={<SaveIcon />}
                    onClick={handleSaveGeneral}
                    disabled={loading}
                  >
                    Save Changes
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </Box>
        )}

        {/* Pricing Plans Tab */}
        {tabValue === 1 && (
          <Box p={3}>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
              <Typography variant="h6">
                Pricing Plans
              </Typography>
              <Button
                variant="outlined"
                startIcon={<AddIcon />}
                onClick={() => setPlanDialog({
                  open: true,
                  isNew: true,
                  plan: {
                    name: '',
                    type: 'free',
                    price_per_call: 0,
                    monthly_price: 0,
                    call_limit: 1000,
                    rate_limit_per_minute: 60,
                    rate_limit_per_day: 10000,
                    is_active: true
                  }
                })}
              >
                Add Plan
              </Button>
            </Box>

            <Grid container spacing={2}>
              {pricingPlans.map((plan) => (
                <Grid item xs={12} md={6} key={plan.id}>
                  <Card>
                    <CardContent>
                      <Box display="flex" justifyContent="space-between" alignItems="start">
                        <Box>
                          <Typography variant="h6" gutterBottom>
                            {plan.name}
                          </Typography>
                          <Chip
                            label={PRICING_TYPES[plan.type]}
                            size="small"
                            color={plan.type === 'free' ? 'default' : 'primary'}
                          />
                        </Box>
                        <Box>
                          <IconButton
                            size="small"
                            onClick={() => setPlanDialog({ open: true, plan, isNew: false })}
                          >
                            <SaveIcon />
                          </IconButton>
                          <IconButton
                            size="small"
                            onClick={() => handleDeletePlan(plan.id)}
                            disabled={pricingPlans.length === 1}
                          >
                            <DeleteIcon />
                          </IconButton>
                        </Box>
                      </Box>

                      <Box mt={2}>
                        {plan.type === 'pay_per_use' && (
                          <Typography variant="h4">
                            ${plan.price_per_call}
                            <Typography variant="caption" color="text.secondary">
                              /call
                            </Typography>
                          </Typography>
                        )}
                        {plan.type === 'subscription' && (
                          <Typography variant="h4">
                            ${plan.monthly_price}
                            <Typography variant="caption" color="text.secondary">
                              /month
                            </Typography>
                          </Typography>
                        )}
                        {plan.type === 'free' && (
                          <Typography variant="h4">Free</Typography>
                        )}
                      </Box>

                      <List dense>
                        <ListItem>
                          <ListItemText
                            primary={`${plan.call_limit ? plan.call_limit.toLocaleString() : 'Unlimited'} calls/month`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemText
                            primary={`${plan.rate_limit_per_minute} requests/minute`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemText
                            primary={`${plan.rate_limit_per_day.toLocaleString()} requests/day`}
                          />
                        </ListItem>
                      </List>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>
          </Box>
        )}

        {/* Documentation Tab */}
        {tabValue === 2 && (
          <Box p={3}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <Card>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      OpenAPI Specification
                    </Typography>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Upload your OpenAPI (Swagger) specification file to enable interactive documentation
                    </Typography>
                    <Box mt={2}>
                      <input
                        accept=".json,.yaml,.yml"
                        style={{ display: 'none' }}
                        id="openapi-file"
                        type="file"
                        onChange={(e) => handleFileUpload(e, 'openapi')}
                      />
                      <label htmlFor="openapi-file">
                        <Button
                          variant="outlined"
                          component="span"
                          startIcon={<UploadIcon />}
                        >
                          Upload OpenAPI Spec
                        </Button>
                      </label>
                      {openApiSpec && (
                        <Chip
                          label="OpenAPI spec uploaded"
                          color="success"
                          size="small"
                          sx={{ ml: 2 }}
                        />
                      )}
                    </Box>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12}>
                <Card>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      Additional Documentation
                    </Typography>
                    <TextField
                      fullWidth
                      multiline
                      rows={10}
                      value={markdownDocs}
                      onChange={(e) => setMarkdownDocs(e.target.value)}
                      placeholder="Add additional documentation in Markdown format..."
                      sx={{ fontFamily: 'monospace' }}
                    />
                    <Box mt={2} display="flex" justifyContent="flex-end">
                      <Button
                        variant="contained"
                        startIcon={<SaveIcon />}
                        onClick={handleSaveGeneral}
                      >
                        Save Documentation
                      </Button>
                    </Box>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          </Box>
        )}
      </Paper>

      {/* Pricing Plan Dialog */}
      <Dialog
        open={planDialog.open}
        onClose={() => setPlanDialog({ open: false, plan: null, isNew: false })}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          {planDialog.isNew ? 'Add Pricing Plan' : 'Edit Pricing Plan'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Plan Name"
                value={planDialog.plan?.name || ''}
                onChange={(e) => setPlanDialog({
                  ...planDialog,
                  plan: { ...planDialog.plan, name: e.target.value }
                })}
              />
            </Grid>
            <Grid item xs={12}>
              <FormControl fullWidth>
                <InputLabel>Pricing Type</InputLabel>
                <Select
                  value={planDialog.plan?.type || 'free'}
                  onChange={(e) => setPlanDialog({
                    ...planDialog,
                    plan: { ...planDialog.plan, type: e.target.value }
                  })}
                  label="Pricing Type"
                >
                  {Object.entries(PRICING_TYPES).map(([value, label]) => (
                    <MenuItem key={value} value={value}>{label}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            {planDialog.plan?.type === 'pay_per_use' && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Price per Call ($)"
                  type="number"
                  inputProps={{ step: 0.001 }}
                  value={planDialog.plan?.price_per_call || 0}
                  onChange={(e) => {
                    const value = parseFloat(e.target.value);
                    if (value >= 0) {
                      setPlanDialog({
                        ...planDialog,
                        plan: { ...planDialog.plan, price_per_call: value }
                      });
                    }
                  }}
                />
              </Grid>
            )}
            {planDialog.plan?.type === 'subscription' && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Monthly Price ($)"
                  type="number"
                  value={planDialog.plan?.monthly_price || 0}
                  onChange={(e) => {
                    const value = parseFloat(e.target.value);
                    if (value >= 0) {
                      setPlanDialog({
                        ...planDialog,
                        plan: { ...planDialog.plan, monthly_price: value }
                      });
                    }
                  }}
                />
              </Grid>
            )}
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Monthly Call Limit"
                type="number"
                value={planDialog.plan?.call_limit || 0}
                onChange={(e) => setPlanDialog({
                  ...planDialog,
                  plan: { ...planDialog.plan, call_limit: parseInt(e.target.value) }
                })}
                helperText="0 for unlimited"
              />
            </Grid>
            <Grid item xs={6}>
              <TextField
                fullWidth
                label="Rate Limit (per minute)"
                type="number"
                value={planDialog.plan?.rate_limit_per_minute || 60}
                onChange={(e) => setPlanDialog({
                  ...planDialog,
                  plan: { ...planDialog.plan, rate_limit_per_minute: parseInt(e.target.value) }
                })}
              />
            </Grid>
            <Grid item xs={6}>
              <TextField
                fullWidth
                label="Rate Limit (per day)"
                type="number"
                value={planDialog.plan?.rate_limit_per_day || 10000}
                onChange={(e) => setPlanDialog({
                  ...planDialog,
                  plan: { ...planDialog.plan, rate_limit_per_day: parseInt(e.target.value) }
                })}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPlanDialog({ open: false, plan: null, isNew: false })}>
            Cancel
          </Button>
          <Button onClick={handleSavePlan} variant="contained">
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default MarketplaceSettings;
