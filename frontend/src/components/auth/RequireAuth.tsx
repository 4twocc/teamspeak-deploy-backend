import React from "react";
import { Navigate } from "react-router-dom";

// children 是被保护的页面组件
const RequireAuth: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem("token");
  // 如果已登录，渲染子组件，否则跳转到 /login
  return token ? <>{children}</> : <Navigate to="/login" replace />;
};

export default RequireAuth;
