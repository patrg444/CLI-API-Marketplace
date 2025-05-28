import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Grid,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Tooltip,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Add as AddIcon,
  Visibility as ViewIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  PlayArrow as DeployIcon,
  Stop as StopIcon,
  Store as MarketplaceIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const mockAPIs = [
  {
    id: 'api-1',
    name: 'Weather API',
    status: 'running',
    version: 'v1.2.3',
    endpoint: 'https://api.api-direct.io/apis/weather-api',
    created: '2024-01-15',
    lastDeployed: '2024-01-20',
    calls: 15234,
    marketplace: true,
  },
  {
    id: 'api-2',
    name: 'Translation Service',
    status: 'stopped',
    version: 'v2.0.1',
    endpoint: 'https://api.api-direct.io/apis/translation-service',
    created: '2024-01-10',
    lastDeployed: '2024-01-18',
    calls: 8921,
    marketplace: false,
  },
];

function APIs() {
  const navigate = useNavigate();
  const [apis, setApis] = useState([]);
  const [loading, setLoading] = useState(true);
  const [deleteDialog, setDeleteDialog] = useState({ open: false, api: null });
  const [copySuccess, setCopySuccess] = useState('');

  useEffect(() => {
    fetchAPIs();
  }, []);

  const fetchAPIs = async () => {
    try {
      // In production, this would call the actual API
      // const response = await axios.get('/api/v1/apis');
      // setApis(response.data);
      
      // For now, use mock data
      setTimeout(() => {
        setApis(mockAPIs);
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to fetch APIs:', error);
      setLoading(false);
    }
  };

  const handleDelete = async (api) => {
    try {
      // await axios.delete(`/api/v1/apis/${api.id}`);
      setApis(apis.filter((a) => a.id !== api.id));
      setDeleteDialog({ open: false, api: null });
    } catch (error) {
      console.error('Failed to delete API:', error);
    }
  };

  const handleStatusToggle = async (api) => {
    try {
      // Toggle status
      const newStatus = api.status === 'running' ? 'stopped' : 'running';
      // await axios.put(`/api/v1/apis/${api.id}/status`, { status: newStatus });
      
      setApis(apis.map((a) => 
        a.id === api.id ? { ...a, status: newStatus } : a
      ));
    } catch (error) {
      console.error('Failed to update API status:', error);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setCopySuccess(text);
    setTimeout(() => setCopySuccess(''), 2000);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'running':
        return 'success';
      case 'stopped':
        return 'default';
      case 'deploying':
        return 'warning';
      default:
        return 'error';
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">
          My APIs
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => window.open('https://docs.api-direct.io/cli/getting-started', '_blank')}
        >
          Deploy New API
        </Button>
      </Box>

      {apis.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No APIs deployed yet
            </Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>
              Get started by deploying your first API using the CLI
            </Typography>
            <Button
              variant="outlined"
              onClick={() => window.open('https://docs.api-direct.io/cli/getting-started', '_blank')}
            >
              View CLI Documentation
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Endpoint</TableCell>
                <TableCell>Last Deployed</TableCell>
                <TableCell align="right">API Calls</TableCell>
                <TableCell>Marketplace</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {apis.map((api) => (
                <TableRow key={api.id}>
                  <TableCell>
                    <Typography variant="subtitle2" fontWeight="bold">
                      {api.name}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={api.status}
                      color={getStatusColor(api.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{api.version}</TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center">
                      <Typography variant="body2" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {api.endpoint}
                      </Typography>
                      <Tooltip title={copySuccess === api.endpoint ? 'Copied!' : 'Copy endpoint'}>
                        <IconButton size="small" onClick={() => copyToClipboard(api.endpoint)}>
                          <CopyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell>{new Date(api.lastDeployed).toLocaleDateString()}</TableCell>
                  <TableCell align="right">{api.calls.toLocaleString()}</TableCell>
                  <TableCell>
                    {api.marketplace ? (
                      <Chip label="Published" color="primary" size="small" icon={<MarketplaceIcon />} />
                    ) : (
                      <Chip label="Private" size="small" variant="outlined" />
                    )}
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="View details">
                      <IconButton size="small" onClick={() => navigate(`/apis/${api.id}`)}>
                        <ViewIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={api.status === 'running' ? 'Stop API' : 'Start API'}>
                      <IconButton size="small" onClick={() => handleStatusToggle(api)}>
                        {api.status === 'running' ? <StopIcon /> : <DeployIcon />}
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Marketplace settings">
                      <IconButton size="small" onClick={() => navigate(`/apis/${api.id}/marketplace`)}>
                        <MarketplaceIcon />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete API">
                      <IconButton 
                        size="small" 
                        onClick={() => setDeleteDialog({ open: true, api })}
                        disabled={api.status === 'running'}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={deleteDialog.open} onClose={() => setDeleteDialog({ open: false, api: null })}>
        <DialogTitle>Delete API</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete "{deleteDialog.api?.name}"? This action cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialog({ open: false, api: null })}>Cancel</Button>
          <Button onClick={() => handleDelete(deleteDialog.api)} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default APIs;
