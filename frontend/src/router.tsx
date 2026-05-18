import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Loading } from './components/common/Loading';
import { useAuthStore } from './store/authStore';

// 懒加载页面组件
const Index = lazy(() => import('./pages/Index').then(m => ({ default: m.Index })));
const Login = lazy(() => import('./pages/Auth/Login').then(m => ({ default: m.Login })));
const Register = lazy(() => import('./pages/Auth/Register').then(m => ({ default: m.Register })));
const GuestEntry = lazy(() => import('./pages/Auth/GuestEntry').then(m => ({ default: m.GuestEntry })));
const ResumeForm = lazy(() => import('./pages/Resume/ResumeForm').then(m => ({ default: m.ResumeForm })));
const PositionSelect = lazy(() => import('./pages/Config/PositionSelect').then(m => ({ default: m.PositionSelect })));
const DirectionSelect = lazy(() => import('./pages/Config/DirectionSelect').then(m => ({ default: m.DirectionSelect })));
const DeviceCheck = lazy(() => import('./pages/Prepare/DeviceCheck').then(m => ({ default: m.DeviceCheck })));
const InterviewRoom = lazy(() => import('./pages/Interview/InterviewRoom').then(m => ({ default: m.InterviewRoom })));
const QuestionnairePage = lazy(() => import('./pages/Questionnaire/QuestionnairePage').then(m => ({ default: m.QuestionnairePage })));
const ReportView = lazy(() => import('./pages/Report/ReportView').then(m => ({ default: m.ReportView })));
const EndPage = lazy(() => import('./pages/End/EndPage').then(m => ({ default: m.EndPage })));


// Suspense 包装器
const SuspenseWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<Loading size="lg" text="加载中..." fullScreen />}>
    {children}
  </Suspense>
);

// 路由守卫：需要认证
const RequireAuth = ({ children }: { children: React.ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
};

// 路由守卫：已认证用户重定向
const RedirectIfAuth = ({ children }: { children: React.ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (token) return <Navigate to="/resume" replace />;
  return <>{children}</>;
};

// 路由配置
export const router = createBrowserRouter([
  {
    path: '/',
    element: (
      <SuspenseWrapper>
        <Index />
      </SuspenseWrapper>
    ),
  },
  {
    path: '/login',
    element: (
      <SuspenseWrapper>
        <RedirectIfAuth>
          <Login />
        </RedirectIfAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/register',
    element: (
      <SuspenseWrapper>
        <RedirectIfAuth>
          <Register />
        </RedirectIfAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/guest',
    element: (
      <SuspenseWrapper>
        <GuestEntry />
      </SuspenseWrapper>
    ),
  },
  {
    path: '/resume',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <ResumeForm />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/config',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <PositionSelect />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/config/direction',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <DirectionSelect />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/prepare',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <DeviceCheck />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/interview',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <InterviewRoom />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/questionnaire',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <QuestionnairePage />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/report',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <ReportView />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '/end',
    element: (
      <SuspenseWrapper>
        <RequireAuth>
          <EndPage />
        </RequireAuth>
      </SuspenseWrapper>
    ),
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

