import { Navigate, Route, Routes } from "react-router-dom";
import { ProtectedRoute } from "@/auth/ProtectedRoute";
import { AdminShell } from "@/components/layout/AdminShell";
import { LoginPage } from "@/pages/LoginPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { UploadsPage } from "@/pages/UploadsPage";
import { UploadNewPage } from "@/pages/UploadNewPage";
import { UploadDetailPage } from "@/pages/UploadDetailPage";
import { DraftReviewPage } from "@/pages/DraftReviewPage";
import { NotFoundPage } from "@/pages/NotFoundPage";

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AdminShell />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/uploads" element={<UploadsPage />} />
          <Route path="/uploads/new" element={<UploadNewPage />} />
          <Route path="/uploads/:assetId" element={<UploadDetailPage />} />
          <Route path="/uploads/:assetId/drafts" element={<DraftReviewPage />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
