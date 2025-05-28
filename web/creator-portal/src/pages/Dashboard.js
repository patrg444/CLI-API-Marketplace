import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  LinearProgress,
  Paper,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  TrendingUp as TrendingUpIcon,
  Api as ApiIcon,
  AttachMoney as MoneyIcon,
  People as PeopleIcon,
  Schedule as ScheduleIcon,
  ArrowForward as ArrowForwardIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler,
} from 'chart.js';
import { useNavigate } from 'react-router-dom';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  ChartTooltip,
  Legend,
  Filler
);

const StatCard = ({ title, value, subtitle, icon, color, trend }) => (
  <Card>
    <CardContent>
      <Box display="flex" justifyContent="space-between" alignItems="center">
        <Box>
          <Typography color="text.secondary" gutterBottom variant="body2">
            {title}
          </Typography>
          <Typography variant="h4" fontWeight="bold">
            {value}
          </Typography>
          {subtitle && (
            <Typography variant="body2" color="text.secondary">
              {subtitle}
            </Typography>
          )}
          {trend && (
            <Box display="flex" alignItems="center" mt={1}>
              <TrendingUpIcon fontSize="small" sx={{ color: 'success.main', mr: 0.5 }} />
              <Typography variant="body2" color="success.main">
                {trend}
              </Typography>
            </Box>
          )}
        </Box>
        <Box
          sx={{
            backgroundColor: `${color}.light`,
            borderRadius: 2,
            p: 1.5,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {React.cloneElement(icon, { sx: { color: `${color}.main`, fontSize: 30 } })}
        </Box>
      </Box>
    </CardContent>
  </Card>
);

function Dashboard() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    totalAPIs: 3,
    totalCalls: 124567,
    revenue: 1847.23,
    activeConsumers: 42,
  });

  const chartData = {
    labels: ['Jan 21', 'Jan 22', 'Jan 23', 'Jan 24', 'Jan 25', 'Jan 26', 'Jan 27'],
    datasets: [
      {
        label: 'API Calls',
        data: [12500, 15200, 18300, 16700, 19800, 21200, 24567],
        borderColor: 'rgb(33, 150, 243)',
        backgroundColor: 'rgba(33, 150, 243, 0.1)',
        tension: 0.4,
        fill: true,
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        mode: 'index',
        intersect: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {
          display: true,
          drawBorder: false,
        },
      },
      x: {
        grid: {
          display: false,
        },
      },
    },
  };

  const recentActivity = [
    {
      type: 'deployment',
      message: 'Weather API v1.2.3 deployed successfully',
      time: '2 hours ago',
      status: 'success',
    },
    {
      type: 'marketplace',
      message: 'Translation Service published to marketplace',
      time: '5 hours ago',
      status: 'success',
    },
    {
      type: 'alert',
      message: 'High traffic detected on Weather API',
      time: '1 day ago',
      status: 'warning',
    },
    {
      type: 'billing',
      message: 'Monthly payout of $1,234.56 processed',
      time: '3 days ago',
      status: 'success',
    },
  ];

  const topAPIs = [
    { name: 'Weather API', calls: 67234, revenue: 892.45 },
    { name: 'Translation Service', calls: 45123, revenue: 623.78 },
    { name: 'Image Processing API', calls: 12210, revenue: 331.00 },
  ];

  useEffect(() => {
    // Simulate loading data
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, []);

  return (
    <Box>
      <Typography variant="h4" fontWeight="bold" mb={3}>
        Dashboard
      </Typography>

      <Grid container spacing={3}>
        {/* Stats Cards */}
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Total APIs"
            value={stats.totalAPIs}
            subtitle="Active APIs"
            icon={<ApiIcon />}
            color="primary"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="API Calls"
            value={stats.totalCalls.toLocaleString()}
            subtitle="This month"
            icon={<TrendingUpIcon />}
            color="success"
            trend="+18.2%"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Revenue"
            value={`$${stats.revenue.toLocaleString()}`}
            subtitle="This month"
            icon={<MoneyIcon />}
            color="warning"
            trend="+12.5%"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Active Consumers"
            value={stats.activeConsumers}
            subtitle="Unique users"
            icon={<PeopleIcon />}
            color="info"
            trend="+5"
          />
        </Grid>

        {/* API Calls Chart */}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">API Calls Trend</Typography>
                <Typography variant="body2" color="text.secondary">
                  Last 7 days
                </Typography>
              </Box>
              <Box height={300}>
                <Line data={chartData} options={chartOptions} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Top APIs */}
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">Top APIs</Typography>
                <IconButton size="small" onClick={() => navigate('/apis')}>
                  <ArrowForwardIcon />
                </IconButton>
              </Box>
              <List>
                {topAPIs.map((api, index) => (
                  <ListItem key={index} divider={index < topAPIs.length - 1}>
                    <ListItemText
                      primary={api.name}
                      secondary={`${api.calls.toLocaleString()} calls`}
                    />
                    <Typography variant="subtitle2" color="success.main">
                      ${api.revenue}
                    </Typography>
                  </ListItem>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>

        {/* Recent Activity */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" mb={2}>
                Recent Activity
              </Typography>
              <List>
                {recentActivity.map((activity, index) => (
                  <ListItem key={index} divider={index < recentActivity.length - 1}>
                    <ListItemIcon>
                      {activity.status === 'success' ? (
                        <CheckCircleIcon color="success" />
                      ) : (
                        <WarningIcon color="warning" />
                      )}
                    </ListItemIcon>
                    <ListItemText
                      primary={activity.message}
                      secondary={
                        <Box display="flex" alignItems="center" gap={1}>
                          <ScheduleIcon fontSize="small" sx={{ color: 'text.secondary' }} />
                          <Typography variant="caption" color="text.secondary">
                            {activity.time}
                          </Typography>
                        </Box>
                      }
                    />
                  </ListItem>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}

export default Dashboard;
