import Login from "@/app/login/Login"
import OverView from "@/app/main/dashboard/overview"
import Approval from "@/app/main/labbook/approval"
import MainLayout from "@/app/main/MainLayout"
import Resource from "@/app/main/resource"
import Schedule from "@/app/main/schedule"
import NotFound from "@/components/not-found"
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  Outlet,
} from "react-router-dom"
import History from "@/app/main/labbook/history"
import Upload from "@/app/main/labbook/upload"
import All from "@/app/main/user/all"
import Register from "@/app/main/user/register"
import Permission from "@/app/main/user/permission"
import General from "@/app/main/setting/general"
import Security from "@/app/main/setting/security"
import Notification from "@/app/main/setting/notification"
import Integration from "@/app/main/setting/integration"
import Signup from "@/app/signout/Signup"

// Pages

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="dashboard/overview" />} />

          <Route path="dashboard" element={<Outlet />}>
            <Route index element={<Navigate to="overview" />} />
            <Route path="overview" element={<OverView />} />
          </Route>
          <Route path="labbook" element={<Outlet />}>
            <Route index element={<Navigate to="upload" />} />
            <Route path="upload" element={<Upload />} />
            <Route path="approval" element={<Approval />} />
            <Route path="history" element={<History />} />
          </Route>
          <Route path="schedule" element={<Schedule />} />
          <Route path="resource" element={<Resource />} />
          <Route path="user" element={<Outlet />}>
            <Route index element={<Navigate to="all" />} />
            <Route path="all" element={<All />} />
            <Route path="register" element={<Register />} />
            <Route path="permission" element={<Permission />} />
          </Route>
          <Route path="setting" element={<Outlet />}>
            <Route index element={<Navigate to="general" />} />
            <Route path="general" element={<General />} />
            <Route path="security" element={<Security />} />
            <Route path="notification" element={<Notification />} />
            <Route path="integration" element={<Integration />} />
          </Route>
        </Route>
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  )
}
