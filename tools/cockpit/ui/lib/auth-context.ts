// auth-context.ts — User role and permissions

export type UserRole = 'developer' | 'devops' | 'admin';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  projects: string[]; // project IDs user has access to
}

export const ROLE_PERMISSIONS: Record<UserRole, string[]> = {
  developer: [
    'view:own_stories',
    'view:own_branch',
    'edit:own_story_spec',
    'create:story',
    'view:capabilities_dev',
    'merge:pull_request',
  ],
  devops: [
    'view:all_stories',
    'view:all_branches',
    'view:all_environments',
    'merge:staging',
    'merge:main',
    'deploy:staging',
    'deploy:production',
    'view:deployment_logs',
    'view:capabilities_prod',
  ],
  admin: [
    '*', // All permissions
  ],
};

export function hasPermission(role: UserRole, permission: string): boolean {
  if (role === 'admin') return true;
  return ROLE_PERMISSIONS[role]?.includes(permission) || false;
}

export function getCurrentUser(): User | null {
  // In a real app, this would come from session/auth provider
  // For now, read from environment or localStorage (for development)
  if (typeof window === 'undefined') return null;

  const stored = localStorage.getItem('cockpit_user');
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      return null;
    }
  }

  return null;
}

export function setCurrentUser(user: User) {
  if (typeof window !== 'undefined') {
    localStorage.setItem('cockpit_user', JSON.stringify(user));
  }
}
