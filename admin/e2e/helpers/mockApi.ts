import type { Page, Route } from "@playwright/test";

/**
 * Network-layer backend mock for the admin app E2E smoke tests.
 *
 * Intercepts every api/v1 request and the S3-style presigned PUT, so the suite
 * runs deterministically with no live server. A minimal in-memory state machine
 * backs the upload -> draft -> publish flow.
 */

// A non-expired admin JWT (exp far in the future). Signature is irrelevant —
// the client only base64-decodes the payload.
function makeAdminJwt(): string {
  const header = Buffer.from(JSON.stringify({ alg: "HS256", typ: "JWT" })).toString("base64url");
  const payload = Buffer.from(
    JSON.stringify({ sub: "admin-1", role: "ADMIN", exp: Math.floor(Date.now() / 1000) + 3600 }),
  ).toString("base64url");
  return `${header}.${payload}.sig`;
}

function ok(route: Route, data: unknown, meta?: unknown) {
  return route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify(meta ? { success: true, data, meta } : { success: true, data }),
  });
}

function fail(route: Route, status: number, code: string, message: string) {
  return route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify({ success: false, data: null, error: { code, message } }),
  });
}

interface MockState {
  asset: Record<string, unknown>;
  parseJobs: unknown[];
  drafts: Record<string, unknown>[];
}

export const ADMIN_TOKEN = makeAdminJwt();

export async function mockBackend(page: Page): Promise<void> {
  const assetId = "asset-1";
  const state: MockState = {
    asset: {
      id: assetId,
      object_key: "uploads/asset-1.pdf",
      bucket_name: "edu",
      original_filename: "de-thi.pdf",
      content_type: "application/pdf",
      file_size: 12345,
      checksum_sha256: "abc123",
      status: "UPLOADED",
      uploaded_by: "admin-1",
      entity_type: "",
      entity_id: "",
      storage_provider: "s3",
      created_at: new Date().toISOString(),
      verified_at: null,
      deleted_at: null,
    },
    parseJobs: [
      {
        id: "job-1",
        asset_id: assetId,
        status: "PARSED",
        parser_version: "v1",
        attempt_count: 1,
        error_message: "",
        claimed_at: null,
        started_at: null,
        finished_at: null,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ],
    drafts: [
      {
        id: "draft-1",
        asset_id: assetId,
        parse_job_id: "job-1",
        question_number: 1,
        question_type: "MCQ",
        stem: "2 + 2 = ?",
        explanation_raw: "Cộng hai số.",
        answer_key_raw: "B",
        parse_confidence: 0.92,
        status: "DRAFT",
        reviewed_by: "",
        reviewed_at: null,
        published_question_id: "",
        options: [
          { id: "opt-a", question_draft_id: "draft-1", option_label: "A", option_text: "3", is_correct_inferred: false, order_index: 0 },
          { id: "opt-b", question_draft_id: "draft-1", option_label: "B", option_text: "4", is_correct_inferred: true, order_index: 1 },
        ],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ],
  };

  // Presigned storage PUT (not under /api/v1).
  await page.route("**/storage-upload/**", (route) => route.fulfill({ status: 200, body: "" }));

  await page.route("**/api/v1/**", async (route) => {
    const req = route.request();
    const url = new URL(req.url());
    const path = url.pathname.replace(/^.*\/api\/v1/, "");
    const method = req.method();

    // --- Auth ---
    if (path === "/auth/login" && method === "POST") {
      const body = req.postDataJSON() as { email?: string; password?: string };
      if (body?.email === "admin@example.com" && body?.password === "password1") {
        return ok(route, { access_token: ADMIN_TOKEN, refresh_token: "refresh-1", expires_at: "" });
      }
      return fail(route, 401, "UNAUTHORIZED", "Sai thông tin đăng nhập");
    }
    if (path === "/auth/logout" && method === "POST") return ok(route, {});
    if (path === "/auth/refresh" && method === "POST") {
      return ok(route, { access_token: ADMIN_TOKEN, refresh_token: "refresh-2", expires_at: "" });
    }

    // --- Uploads ---
    if (path === "/admin/uploads/init" && method === "POST") {
      return ok(route, {
        asset_id: assetId,
        object_key: "uploads/asset-1.pdf",
        upload_url: `${url.origin}/storage-upload/asset-1.pdf`,
        method: "PUT",
        headers: { "Content-Type": "application/pdf" },
        expires_at: "",
      });
    }
    if (path === "/admin/uploads/complete" && method === "POST") {
      return ok(route, { asset: state.asset, parse_job_id: "job-1" });
    }
    if (path === "/admin/uploads" && method === "GET") {
      return ok(route, [state.asset], { total: 1, page: 1, limit: 20 });
    }
    if (path === `/admin/uploads/${assetId}` && method === "GET") return ok(route, state.asset);
    if (path === `/admin/uploads/${assetId}` && method === "DELETE") return ok(route, {});
    if (path === `/admin/uploads/${assetId}/parse-jobs` && method === "GET") return ok(route, state.parseJobs);
    if (path === `/admin/uploads/${assetId}/draft-questions` && method === "GET") return ok(route, state.drafts);
    if (path === `/admin/uploads/${assetId}/publish` && method === "POST") {
      state.drafts = state.drafts.map((d) => ({ ...d, status: "PUBLISHED" }));
      return ok(route, { published: state.drafts.length });
    }

    // --- Drafts ---
    if (/^\/admin\/question-drafts\/[^/]+\/publish$/.test(path) && method === "POST") {
      state.drafts = state.drafts.map((d) => ({ ...d, status: "PUBLISHED" }));
      return ok(route, { id: "q-1" });
    }
    if (/^\/admin\/question-drafts\/[^/]+\/options\/[^/]+$/.test(path) && method === "PUT") return ok(route, {});
    if (/^\/admin\/question-drafts\/[^/]+$/.test(path) && method === "PUT") return ok(route, {});

    // --- Curriculum (for topic selector) ---
    if (path === "/curriculum/subjects" && method === "GET") {
      return ok(route, [{ id: "subj-1", code: "MATH", name: "Toán", grade_level: 10 }]);
    }
    if (/^\/curriculum\/subjects\/[^/]+\/chapters$/.test(path) && method === "GET") {
      return ok(route, [{ id: "chap-1", subject_id: "subj-1", title: "Chương 1", order_index: 0 }]);
    }
    if (/^\/curriculum\/chapters\/[^/]+\/topics$/.test(path) && method === "GET") {
      return ok(route, [{ id: "topic-1", chapter_id: "chap-1", title: "Chủ đề 1", order_index: 0 }]);
    }

    // Default: empty success so the UI never hangs on an unhandled call.
    return ok(route, []);
  });
}

/** Seeds sessionStorage so the app boots already authenticated as ADMIN. */
export async function seedAuth(page: Page): Promise<void> {
  await page.addInitScript((token: string) => {
    sessionStorage.setItem("edu_admin_access", token);
    sessionStorage.setItem("edu_admin_refresh", "refresh-1");
  }, ADMIN_TOKEN);
}
