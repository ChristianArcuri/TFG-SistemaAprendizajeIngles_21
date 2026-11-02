import api from './client'

export async function register({ email, password, fullName }) {
  const { data } = await api.post('/api/v1/auth/register', {
    email,
    password,
    fullName,
  })
  return data.data
}

export async function login({ email, password }) {
  const { data } = await api.post('/api/v1/auth/login', { email, password })
  return data.data
}

export async function fetchMe() {
  const { data } = await api.get('/api/v1/me')
  return data.data
}

export async function updateProfile(profile) {
  const { data } = await api.put('/api/v1/me/profile', profile)
  return data.data
}
