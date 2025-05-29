import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Stepper,
  Step,
  StepLabel,
  Alert,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
  Card,
  CardContent,
  Divider,
} from '@mui/material';
import {
  Check as CheckIcon,
  AccountCircle as AccountIcon,
  Business as BusinessIcon,
  CreditCard as CardIcon,
  Security as SecurityIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';

const steps = ['Create Account', 'Verify Identity', 'Add Bank Details', 'Review & Submit'];

const statusMessages = {
  not_connected: 'You haven\'t connected your Stripe account yet. Click below to get started.',
  pending: 'Your account is being reviewed by Stripe. This usually takes 1-3 business days.',
  requires_information: 'Additional information is required to complete your account setup.',
  active: 'Your Stripe account is active and ready to receive payouts!',
  rejected: 'Your account application was rejected. Please contact support for assistance.',
};

function StripeConnectOnboarding({ accountStatus, onStatusUpdate }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const getActiveStep = () => {
    if (!accountStatus || accountStatus.status === 'not_connected') return 0;
    if (accountStatus.status === 'pending') return 2;
    if (accountStatus.status === 'requires_information') return 1;
    if (accountStatus.status === 'active') return 4;
    return 0;
  };

  const handleConnectStripe = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`${process.env.REACT_APP_API_URL}/payout/accounts/connect`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          return_url: `${window.location.origin}/payouts?stripe_connect=success`,
          refresh_url: `${window.location.origin}/payouts?stripe_connect=refresh`,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to create Stripe Connect link');
      }

      const data = await response.json();
      
      // Redirect to Stripe Connect onboarding
      window.location.href = data.url;
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRefreshUrl = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`${process.env.REACT_APP_API_URL}/payout/accounts/refresh-url`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to refresh onboarding link');
      }

      const data = await response.json();
      window.location.href = data.url;
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const renderAccountStatus = () => {
    if (!accountStatus) return null;

    const statusColor = {
      active: 'success',
      pending: 'warning',
      requires_information: 'warning',
      not_connected: 'info',
      rejected: 'error',
    }[accountStatus.status];

    return (
      <Card sx={{ mb: 4 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6" sx={{ flexGrow: 1 }}>
              Account Status
            </Typography>
            <Alert severity={statusColor} icon={false}>
              {accountStatus.status.replace(/_/g, ' ').toUpperCase()}
            </Alert>
          </Box>
          
          <Typography variant="body2" color="textSecondary" paragraph>
            {statusMessages[accountStatus.status]}
          </Typography>

          {accountStatus.status === 'active' && accountStatus.details && (
            <Box sx={{ mt: 2 }}>
              <Divider sx={{ mb: 2 }} />
              <Typography variant="subtitle2" gutterBottom>
                Account Details
              </Typography>
              <List dense>
                <ListItem>
                  <ListItemText 
                    primary="Account Type"
                    secondary={accountStatus.details.type || 'Individual'}
                  />
                </ListItem>
                <ListItem>
                  <ListItemText 
                    primary="Country"
                    secondary={accountStatus.details.country || 'United States'}
                  />
                </ListItem>
                <ListItem>
                  <ListItemText 
                    primary="Default Currency"
                    secondary={accountStatus.details.default_currency?.toUpperCase() || 'USD'}
                  />
                </ListItem>
              </List>
            </Box>
          )}
        </CardContent>
      </Card>
    );
  };

  return (
    <Box>
      {/* Header */}
      <Typography variant="h5" gutterBottom sx={{ mb: 3 }}>
        Stripe Connect Setup
      </Typography>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Account Status */}
      {renderAccountStatus()}

      {/* Setup Steps */}
      {accountStatus?.status !== 'active' && (
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Setup Progress
          </Typography>
          <Stepper activeStep={getActiveStep()} alternativeLabel sx={{ mb: 4 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>
        </Paper>
      )}

      {/* Requirements Checklist */}
      <Paper sx={{ p: 3, mb: 4 }}>
        <Typography variant="h6" gutterBottom>
          What You'll Need
        </Typography>
        <Typography variant="body2" color="textSecondary" paragraph>
          To complete your Stripe Connect account setup, please have the following ready:
        </Typography>
        
        <List>
          <ListItem>
            <ListItemIcon>
              <AccountIcon color="primary" />
            </ListItemIcon>
            <ListItemText 
              primary="Personal Information"
              secondary="Legal name, date of birth, and address"
            />
          </ListItem>
          <ListItem>
            <ListItemIcon>
              <BusinessIcon color="primary" />
            </ListItemIcon>
            <ListItemText 
              primary="Business Details (if applicable)"
              secondary="Business name, type, and tax ID"
            />
          </ListItem>
          <ListItem>
            <ListItemIcon>
              <CardIcon color="primary" />
            </ListItemIcon>
            <ListItemText 
              primary="Bank Account Information"
              secondary="Account and routing numbers for payouts"
            />
          </ListItem>
          <ListItem>
            <ListItemIcon>
              <SecurityIcon color="primary" />
            </ListItemIcon>
            <ListItemText 
              primary="Identity Verification"
              secondary="Government-issued ID or passport"
            />
          </ListItem>
        </List>
      </Paper>

      {/* Call to Action */}
      <Box sx={{ textAlign: 'center' }}>
        {accountStatus?.status === 'not_connected' && (
          <Button
            variant="contained"
            size="large"
            onClick={handleConnectStripe}
            disabled={loading}
            startIcon={loading ? <CircularProgress size={20} /> : null}
          >
            Connect with Stripe
          </Button>
        )}

        {accountStatus?.status === 'requires_information' && (
          <Button
            variant="contained"
            size="large"
            onClick={handleRefreshUrl}
            disabled={loading}
            startIcon={loading ? <CircularProgress size={20} /> : null}
          >
            Continue Setup
          </Button>
        )}

        {accountStatus?.status === 'active' && (
          <Box>
            <CheckIcon sx={{ fontSize: 48, color: 'success.main', mb: 2 }} />
            <Typography variant="h6" color="success.main">
              Your account is all set!
            </Typography>
            <Typography variant="body2" color="textSecondary">
              You'll receive payouts on the 1st of each month for earnings over $25.
            </Typography>
          </Box>
        )}
      </Box>

      {/* Information Box */}
      <Box sx={{ mt: 4 }}>
        <Alert severity="info">
          <Typography variant="body2">
            <strong>Important:</strong> API-Direct partners with Stripe to handle secure payouts. 
            Your financial information is never stored on our servers. Stripe Connect is required 
            to receive payments and comply with financial regulations.
          </Typography>
        </Alert>
      </Box>
    </Box>
  );
}

export default StripeConnectOnboarding;
