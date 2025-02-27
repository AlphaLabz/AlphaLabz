import * as React from "react"
import {
  CalendarIcon,
  FileTextIcon,
  HomeIcon,
  PackageIcon,
  SettingsIcon,
  UsersIcon,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { AppTitle } from "@/components/app-title"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"

// This is sample data.
const data = {
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: HomeIcon,
      isActive: true,
      items: [
        {
          title: "Overview",
          url: "/dashboard/overview",
        },
      ],
    },
    {
      title: "Lab Books",
      url: "/labbook",
      icon: FileTextIcon,
      items: [
        {
          title: "Upload",
          url: "/labbook/upload",
        },
        {
          title: "Approval",
          url: "/labbook/approval",
        },
        {
          title: "History",
          url: "/labbook/history",
        },
      ],
    },
    {
      title: "Schedule",
      url: "/schedule",
      icon: CalendarIcon,
    },
    {
      title: "Resources",
      url: "/resource",
      icon: PackageIcon, 
    },
    {
      title: "User",
      url: "/user",
      icon: UsersIcon,
      items: [
        {
          title: "All User",
          url: "/user/all",
        },
        {
          title: "Register",
          url: "/user/register",
        },
        {
          title: "Permissions",
          url: "/user/permission",
        },
      ],
    },
    {
      title: "Settings",
      url: "/setting",
      icon: SettingsIcon,
      items: [
        {
          title: "General",
          url: "/setting/general",
        },
        {
          title: "Security",
          url: "/setting/security",
        },
        {
          title: "Notifications",
          url: "/setting/notification",
        },
        {
          title: "Integrations",
          url: "/setting/integration",
        },
      ],
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
       <AppTitle />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
