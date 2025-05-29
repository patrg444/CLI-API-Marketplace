import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
} from '@mui/material';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

function EarningsDashboard() {
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState('6months');
  const [earningsData, setEarningsData] = useState({
    monthly_trends: [],
    api_breakdown: [],
    top_apis: [],
    total_statistics: {
      total_revenue: 0,
      platform_commission: 0,
      net_earnings: 0,
    },
  });

  useEffect(() => {
    fetchEarningsData();
  }, [timeRange]);

  const fetchEarningsData = async () => {
    setLoading(true);
    try {
      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/payout/earnings/analytics?range=${timeRange}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          },
        }
      );
      const data = await response.json();
      setEarningsData(data);
    } catch (error) {
      console.error('Failed to fetch earnings data:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value) => {
    return `$${value.toFixed(2)}`;
  };

  const formatPercent = (value) => {
    return `${value.toFixed(1)}%`;
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      {/* Time Range Selector */}
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'flex-end' }}>
        <FormControl size="small">
          <InputLabel>Time Range</InputLabel>
          <Select
            value={timeRange}
            label="Time Range"
            onChange={(e) => setTimeRange(e.target.value)}
          >
            <MenuItem value="1month">Last Month</MenuItem>
            <MenuItem value="3months">Last 3 Months</MenuItem>
            <MenuItem value="6months">Last 6 Months</MenuItem>
            <MenuItem value="12months">Last Year</MenuItem>
            <MenuItem value="all">All Time</MenuItem>
          </Select>
        </FormControl>
      </Box>

      {/* Statistics Summary */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="body2" color="textSecondary" gutterBottom>
              Total Revenue
            </Typography>
            <Typography variant="h5">
              {formatCurrency(earningsData.total_statistics.total_revenue)}
            </Typography>
            <Typography variant="caption" color="textSecondary">
              From API subscriptions
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="body2" color="textSecondary" gutterBottom>
              Platform Commission (20%)
            </Typography>
            <Typography variant="h5" color="error">
              -{formatCurrency(earningsData.total_statistics.platform_commission)}
            </Typography>
            <Typography variant="caption" color="textSecondary">
              API-Direct platform fee
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="body2" color="textSecondary" gutterBottom>
              Net Earnings
            </Typography>
            <Typography variant="h5" color="success.main">
              {formatCurrency(earningsData.total_statistics.net_earnings)}
            </Typography>
            <Typography variant="caption" color="textSecondary">
              Your total earnings
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Charts */}
      <Grid container spacing={3}>
        {/* Monthly Earnings Trend */}
        <Grid item xs={12} lg={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Monthly Earnings Trend
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={earningsData.monthly_trends}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis tickFormatter={formatCurrency} />
                <Tooltip formatter={formatCurrency} />
                <Legend />
                <Line 
                  type="monotone" 
                  dataKey="gross_revenue" 
                  stroke="#8884d8" 
                  name="Gross Revenue"
                  strokeWidth={2}
                />
                <Line 
                  type="monotone" 
                  dataKey="net_earnings" 
                  stroke="#82ca9d" 
                  name="Net Earnings"
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>

        {/* API Revenue Distribution */}
        <Grid item xs={12} lg={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Revenue by API
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={earningsData.api_breakdown}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${formatPercent(percent * 100)}`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="revenue"
                >
                  {earningsData.api_breakdown.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip formatter={formatCurrency} />
              </PieChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>

        {/* Top Performing APIs */}
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Top Performing APIs
            </Typography>
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>API Name</TableCell>
                    <TableCell align="right">Subscribers</TableCell>
                    <TableCell align="right">Gross Revenue</TableCell>
                    <TableCell align="right">Commission</TableCell>
                    <TableCell align="right">Net Earnings</TableCell>
                    <TableCell align="right">Avg. Revenue/User</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {earningsData.top_apis.map((api) => (
                    <TableRow key={api.api_id}>
                      <TableCell>{api.name}</TableCell>
                      <TableCell align="right">{api.subscriber_count}</TableCell>
                      <TableCell align="right">{formatCurrency(api.gross_revenue)}</TableCell>
                      <TableCell align="right" sx={{ color: 'error.main' }}>
                        -{formatCurrency(api.commission)}
                      </TableCell>
                      <TableCell align="right" sx={{ color: 'success.main' }}>
                        {formatCurrency(api.net_earnings)}
                      </TableCell>
                      <TableCell align="right">
                        {formatCurrency(api.avg_revenue_per_user)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>

        {/* Monthly Comparison Bar Chart */}
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Monthly Revenue Comparison
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={earningsData.monthly_trends}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis tickFormatter={formatCurrency} />
                <Tooltip formatter={formatCurrency} />
                <Legend />
                <Bar dataKey="gross_revenue" fill="#8884d8" name="Gross Revenue" />
                <Bar dataKey="commission" fill="#ff7675" name="Commission" />
                <Bar dataKey="net_earnings" fill="#00b894" name="Net Earnings" />
              </BarChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}

export default EarningsDashboard;
