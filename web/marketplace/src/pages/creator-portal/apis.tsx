import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Layout from '../../components/Layout';

const CreatorAPIs: React.FC = () => {
  const router = useRouter();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [apis, setApis] = useState([
    {
      id: 1,
      name: 'Test Payment API',
      description: 'A test API for payment processing',
      status: 'Published',
      category: 'Financial Services',
      subscribers: 247,
      earnings: 5432.10
    }
  ]);

  const handleCreateAPI = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const newAPI = {
      id: apis.length + 1,
      name: formData.get('name') as string,
      description: formData.get('description') as string,
      category: formData.get('category') as string,
      status: 'Draft',
      subscribers: 0,
      earnings: 0
    };
    
    setApis([...apis, newAPI]);
    setShowCreateModal(false);
    
    // Show success message
    const message = document.createElement('div');
    message.textContent = 'API created successfully!';
    message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
    document.body.appendChild(message);
    setTimeout(() => message.remove(), 3000);
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">My APIs</h1>
          <button
            className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
            data-testid="create-api-button"
            onClick={() => setShowCreateModal(true)}
          >
            Create New API
          </button>
        </div>

        {/* API List */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {apis.map((api) => (
            <div 
              key={api.id} 
              className="bg-white shadow rounded-lg p-6 cursor-pointer hover:shadow-lg transition-shadow"
              data-testid={`api-card-${api.name}`}
              onClick={() => router.push(`/creator-portal/apis/${api.id}`)}
            >
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-lg font-medium text-gray-900">{api.name}</h3>
                <span 
                  className={`px-2 py-1 text-xs rounded-full ${
                    api.status === 'Published' 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-yellow-100 text-yellow-800'
                  }`}
                  data-testid="api-status"
                >
                  {api.status}
                </span>
              </div>
              
              <p className="text-gray-600 text-sm mb-4">{api.description}</p>
              
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-500">Category:</span>
                  <span>{api.category}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Subscribers:</span>
                  <span>{api.subscribers}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Earnings:</span>
                  <span className="font-medium">${api.earnings.toFixed(2)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Create API Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-md w-full">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Create New API</h3>
              
              <form onSubmit={handleCreateAPI} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">API Name</label>
                  <input
                    type="text"
                    name="name"
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="api-name-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                  <textarea
                    name="description"
                    required
                    rows={3}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="api-description-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
                  <select
                    name="category"
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="api-category-select"
                  >
                    <option value="">Select a category</option>
                    <option value="Financial Services">Financial Services</option>
                    <option value="Data & Analytics">Data & Analytics</option>
                    <option value="Communication">Communication</option>
                    <option value="Developer Tools">Developer Tools</option>
                  </select>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">OpenAPI Specification</label>
                  <input
                    type="file"
                    name="openapi"
                    accept=".json,.yaml,.yml"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="openapi-upload"
                  />
                </div>
                
                <div className="flex justify-end space-x-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-500"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700"
                    data-testid="create-api-submit"
                  >
                    Create API
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default CreatorAPIs;