import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'landing',
      component: () => import('../views/LandingView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/onboarding',
      name: 'onboarding',
      component: () => import('../views/OnboardingView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
    {
      path: '/practice/daily/:categoryCode',
      name: 'practice-daily',
      component: () => import('../views/DailyPracticeView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
    {
      path: '/practice/hr-interview',
      name: 'practice-hr-interview',
      component: () => import('../views/HrInterviewPracticeView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
    {
      path: '/quiz',
      name: 'practice-quiz',
      component: () => import('../views/QuizView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
    {
      path: '/results/:sessionId?',
      name: 'results',
      component: () => import('../views/ResultsView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('../views/HistoryView.vue'),
      meta: { requiresAuth: true, requiresProfile: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (auth.isAuthenticated && !auth.user) {
    try {
      await auth.loadMe()
    } catch {
      auth.logout()
    }
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    return auth.hasProfile ? { name: 'dashboard' } : { name: 'onboarding' }
  }

  if (to.meta.requiresProfile && auth.isAuthenticated && !auth.hasProfile) {
    if (to.name !== 'onboarding') {
      return { name: 'onboarding' }
    }
  }

  return true
})

export default router
