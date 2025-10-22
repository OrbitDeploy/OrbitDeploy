export interface Configuration {
  id: number;
  applicationId: number;
  version: number;
  envVars: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateConfigurationRequest {
  version: number;
  envVars: string;
  isActive: boolean;
}

export interface UpdateConfigurationRequest {
  version: number;
  envVars: string;
  isActive?: boolean; // Optional field to update active status
}