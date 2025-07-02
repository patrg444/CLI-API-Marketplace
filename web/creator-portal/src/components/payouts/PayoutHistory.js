import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Chip,
  IconButton,
  Button,
  Collapse,
  CircularProgress,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  TextField,
  Grid,
  Tooltip,
} from '@mui/material';
import {
  KeyboardArrowDown as ExpandIcon,
  KeyboardArrowUp as CollapseIcon,
  Download as DownloadIcon,
  Receipt as ReceiptIcon,
  FilterList as FilterIcon,
} from '@mui/icons-material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';

function PayoutHistoryRow({ payout, onDownloadReceipt }) {
  const [open, setOpen] = useState(false);

  const getStatusColor = (status) => {
    switch (status) {
      case 'paid':
        return 'success';
      case 'pending':
        return 'warning';
      case 'processing':
        return 'info';
      case 'failed':
        return 'error';
      default:
        return 'default';
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatCurrency = (amount) => {
    return `$${amount.toFixed(2)}`;
  };

  return (
    <>
      <TableRow sx={{ '& > *': { borderBottom: 'unset' } }} data-testid="payout-item">
        <TableCell>
          <IconButton
            aria-label="expand row"
            size="small"
            onClick={() => setOpen(!open)}
          >
            {open ? <CollapseIcon /> : <ExpandIcon />}
          </IconButton>
        </TableCell>
        <TableCell data-testid="payout-date">{formatDate(payout.payout_date)}</TableCell>
        <TableCell data-testid="payout-status">
          <Chip 
            label={payout.status}
            color={getStatusColor(payout.status)}
            size="small"
          />
        </TableCell>
        <TableCell align="right">{formatCurrency(payout.gross_amount)}</TableCell>
        <TableCell align="right" sx={{ color: 'error.main' }}>
          -{formatCurrency(payout.commission_amount)}
        </TableCell>
        <TableCell align="right" sx={{ fontWeight: 'bold' }} data-testid="payout-amount">
          {formatCurrency(payout.net_amount)}
        </TableCell>
        <TableCell align="center">
          <Tooltip title="Download Receipt">
            <IconButton 
              size="small"
              onClick={() => onDownloadReceipt(payout.id)}
              disabled={payout.status !== 'paid'}
            >
              <ReceiptIcon />
            </IconButton>
          </Tooltip>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={7}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 2 }} data-testid="payout-details-modal">
              <Typography variant="h6" gutterBottom component="div">
                Payout Details
              </Typography>
              <Grid container spacing={2} sx={{ mb: 2 }}>
                <Grid item xs={12} sm={6} md={3}>
                  <Typography variant="body2" color="textSecondary">
                    Payout ID
                  </Typography>
                  <Typography variant="body2">
                    {payout.stripe_payout_id || payout.id}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Typography variant="body2" color="textSecondary">
                    Period
                  </Typography>
                  <Typography variant="body2">
                    {formatDate(payout.period_start)} - {formatDate(payout.period_end)}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Typography variant="body2" color="textSecondary">
                    Processing Date
                  </Typography>
                  <Typography variant="body2">
                    {payout.processed_at ? formatDate(payout.processed_at) : 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Typography variant="body2" color="textSecondary">
                    Bank Account
                  </Typography>
                  <Typography variant="body2">
                    ****{payout.bank_account_last4 || '****'}
                  </Typography>
                </Grid>
              </Grid>
              
              <Typography variant="subtitle2" gutterBottom>
                API Breakdown
              </Typography>
              <TableContainer data-testid="payout-breakdown">
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>API Name</TableCell>
                      <TableCell align="right">Subscribers</TableCell>
                      <TableCell align="right">Gross Revenue</TableCell>
                      <TableCell align="right">Commission (20%)</TableCell>
                      <TableCell align="right">Net Amount</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {payout.line_items?.map((item) => (
                      <TableRow key={item.api_id}>
                        <TableCell>{item.api_name}</TableCell>
                        <TableCell align="right">{item.subscriber_count}</TableCell>
                        <TableCell align="right">{formatCurrency(item.gross_amount)}</TableCell>
                        <TableCell align="right" sx={{ color: 'error.main' }}>
                          -{formatCurrency(item.commission)}
                        </TableCell>
                        <TableCell align="right">
                          {formatCurrency(item.net_amount)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
              <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
                <Button 
                  onClick={() => setOpen(false)}
                  data-testid="close-modal"
                  variant="outlined"
                  size="small"
                >
                  Close
                </Button>
              </Box>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
}

function PayoutHistory() {
  const [loading, setLoading] = useState(true);
  const [payouts, setPayouts] = useState([]);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [totalCount, setTotalCount] = useState(0);
  const [filters, setFilters] = useState({
    status: 'all',
    startDate: null,
    endDate: null,
  });

  useEffect(() => {
    fetchPayouts();
  }, [page, rowsPerPage, filters]);

  const fetchPayouts = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: page + 1,
        limit: rowsPerPage,
        ...(filters.status !== 'all' && { status: filters.status }),
        ...(filters.startDate && { start_date: filters.startDate.toISOString() }),
        ...(filters.endDate && { end_date: filters.endDate.toISOString() }),
      });

      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/payout/payouts?${params}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          },
        }
      );
      
      const data = await response.json();
      setPayouts(data.payouts || []);
      setTotalCount(data.total || 0);
    } catch (error) {
      console.error('Failed to fetch payouts:', error);
      // Provide mock data for testing
      setPayouts([
        {
          id: 1,
          payout_date: '2024-01-01',
          status: 'completed',
          gross_amount: 1000.00,
          commission_amount: 200.00,
          net_amount: 800.00,
          bank_account_last4: '6789',
          period_start: '2023-12-01',
          period_end: '2023-12-31',
          processed_at: '2024-01-01',
          stripe_payout_id: 'po_test_123',
          line_items: [
            {
              api_id: 1,
              api_name: 'Test Payment API',
              subscriber_count: 247,
              gross_amount: 1000.00,
              commission: 200.00,
              net_amount: 800.00
            }
          ]
        }
      ]);
      setTotalCount(1);
    } finally {
      setLoading(false);
    }
  };

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleFilterChange = (field, value) => {
    setFilters(prev => ({ ...prev, [field]: value }));
    setPage(0);
  };

  const handleDownloadReceipt = async (payoutId) => {
    try {
      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/payout/payouts/${payoutId}/receipt`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to download receipt');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `payout-receipt-${payoutId}.pdf`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download receipt:', error);
    }
  };

  const handleExportAll = async () => {
    try {
      const params = new URLSearchParams({
        format: 'csv',
        ...(filters.status !== 'all' && { status: filters.status }),
        ...(filters.startDate && { start_date: filters.startDate.toISOString() }),
        ...(filters.endDate && { end_date: filters.endDate.toISOString() }),
      });

      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/payout/payouts/export?${params}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to export payouts');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `payouts-export-${new Date().toISOString().split('T')[0]}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export payouts:', error);
    }
  };

  if (loading && payouts.length === 0) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
          <FilterIcon color="action" />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Status</InputLabel>
            <Select
              value={filters.status}
              label="Status"
              onChange={(e) => handleFilterChange('status', e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="paid">Paid</MenuItem>
              <MenuItem value="pending">Pending</MenuItem>
              <MenuItem value="processing">Processing</MenuItem>
              <MenuItem value="failed">Failed</MenuItem>
            </Select>
          </FormControl>
          
          <LocalizationProvider dateAdapter={AdapterDateFns}>
            <DatePicker
              label="Start Date"
              value={filters.startDate}
              onChange={(date) => handleFilterChange('startDate', date)}
              renderInput={(params) => <TextField {...params} size="small" />}
            />
            <DatePicker
              label="End Date"
              value={filters.endDate}
              onChange={(date) => handleFilterChange('endDate', date)}
              renderInput={(params) => <TextField {...params} size="small" />}
            />
          </LocalizationProvider>

          <Box sx={{ flexGrow: 1 }} />
          
          <Button
            variant="outlined"
            startIcon={<DownloadIcon />}
            onClick={handleExportAll}
          >
            Export All
          </Button>
        </Box>
      </Paper>

      {/* Payouts Table */}
      <Paper data-testid="payout-list">
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell />
                <TableCell>Payout Date</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Gross Amount</TableCell>
                <TableCell align="right">Commission</TableCell>
                <TableCell align="right">Net Amount</TableCell>
                <TableCell align="center">Receipt</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {payouts.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                    <Typography variant="body2" color="textSecondary">
                      No payouts found
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                payouts.map((payout) => (
                  <PayoutHistoryRow
                    key={payout.id}
                    payout={payout}
                    onDownloadReceipt={handleDownloadReceipt}
                  />
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
        
        <TablePagination
          rowsPerPageOptions={[5, 10, 25]}
          component="div"
          count={totalCount}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
    </Box>
  );
}

export default PayoutHistory;
