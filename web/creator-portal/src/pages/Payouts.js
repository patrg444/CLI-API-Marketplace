import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  Alert,
  Tab,
  Tabs,
  CircularProgress,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  AttachMoney as MoneyIcon,
  TrendingUp as TrendingIcon,
  AccountBalance as BankIcon,
  Info as InfoIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import EarningsDashboard from '../components/payouts/EarningsDashboard';
import StripeConnectOnboarding from '../components/payouts/StripeConnectOnboarding';
import PayoutHistory from '../components/payouts/PayoutHistory';

function TabPanel({ children, value, index, ...other }) {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`payout-tabpanel-${index}`}
      aria-labelledby={`payout-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
    </div>
  );
}

function Payouts() {
  const [tabValue, setTabValue] = useState(0);
  const [loading, setLoading] = useState(true);
  const [accountStatus, setAccountStatus] = useState(null);
  const [earnings, setEarnings] = useState({
    current_month: 0,
    pending_payout: 0,
    total_earned: 0,
    platform_commission: 0,
  });

  useEffect(() => {
    fetchAccountStatus();
    fetchEarnings();
  }, []);

  const fetchAccountStatus = async () => {
    try {
      const response = await fetch(`${process.env.REACT_APP_API_URL}/payout/accounts/status`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
        },
      });
      const data = await response.json();
      setAccountStatus(data);
    } catch (error) {
      console.error('Failed to fetch account status:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchEarnings = async () => {
    try {
      const response = await fetch(`${process.env.REACT_APP_API_URL}/payout/earnings/current`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
        },
      });
      const data = await response.json();
      setEarnings(data);
    } catch (error) {
      console.error('Failed to fetch earnings:', error);
    }
  };

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const handleRefresh = () => {
    fetchAccountStatus();
    fetchEarnings();
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h4" gutterBottom>
          Payouts
        </Typography>
        <IconButton onClick={handleRefresh} color="primary">
          <RefreshIcon />
        </IconButton>
      </Box>

      {/* Account Status Alert */}
      {accountStatus && accountStatus.status !== 'active' && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {accountStatus.status === 'not_connected' && 
            'Connect your Stripe account to start receiving payouts.'}
          {accountStatus.status === 'pending' && 
            'Your Stripe account is being reviewed. You\'ll be able to receive payouts once approved.'}
          {accountStatus.status === 'requires_information' && 
            'Additional information is required to complete your Stripe account setup.'}
        </Alert>
      )}

      {/* Summary Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <MoneyIcon color="primary" sx={{ mr: 1 }} />
                <Typography color="textSecondary" variant="body2">
                  Current Month
                </Typography>
              </Box>
              <Typography variant="h5" component="div">
                ${earnings.current_month.toFixed(2)}
              </Typography>
              <Typography variant="caption" color="textSecondary">
                Before commission
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <TrendingIcon color="primary" sx={{ mr: 1 }} />
                <Typography color="textSecondary" variant="body2">
                  Pending Payout
                </Typography>
                <Tooltip title="Amount to be paid on the 1st of next month (minimum $25)">
                  <InfoIcon fontSize="small" sx={{ ml: 'auto', color: 'text.secondary' }} />
                </Tooltip>
              </Box>
              <Typography variant="h5" component="div">
                ${earnings.pending_payout.toFixed(2)}
              </Typography>
              <Typography variant="caption" color="textSecondary">
                After 20% commission
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <BankIcon color="primary" sx={{ mr: 1 }} />
                <Typography color="textSecondary" variant="body2">
                  Total Earned
                </Typography>
              </Box>
              <Typography variant="h5" component="div">
                ${earnings.total_earned.toFixed(2)}
              </Typography>
              <Typography variant="caption" color="textSecondary">
                All time earnings
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <Typography color="textSecondary" variant="body2">
                  Account Status
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center' }}>
                <Chip 
                  label={accountStatus?.status || 'Not Connected'}
                  color={accountStatus?.status === 'active' ? 'success' : 'warning'}
                  size="small"
                />
              </Box>
              <Typography variant="caption" color="textSecondary" sx={{ mt: 1 }}>
                Stripe Connect
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Tabs */}
      <Paper sx={{ width: '100%' }}>
        <Tabs 
          value={tabValue} 
          onChange={handleTabChange} 
          aria-label="payout tabs"
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label="Earnings Dashboard" />
          <Tab label="Payout History" />
          <Tab label="Account Setup" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          <EarningsDashboard />
        </TabPanel>
        <TabPanel value={tabValue} index={1}>
          <PayoutHistory />
        </TabPanel>
        <TabPanel value={tabValue} index={2}>
          <StripeConnectOnboarding 
            accountStatus={accountStatus}
            onStatusUpdate={fetchAccountStatus}
          />
        </TabPanel>
      </Paper>
    </Box>
  );
}

export default Payouts;
