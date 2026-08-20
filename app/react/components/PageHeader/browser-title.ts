const DEFAULT_TITLE = 'Portainer';

let environmentName: string | undefined;
let pageContext: string | undefined;
let pageTitle: string | undefined;

export function setBrowserTitleEnvironment(name?: string) {
  environmentName = name;
  updateBrowserTitle();
}

export function setBrowserTitlePage({
  title,
  breadcrumbs,
}: {
  title?: string;
  breadcrumbs: BrowserTitleBreadcrumbInput[] | string;
}) {
  pageContext = getLastBreadcrumbLabel(breadcrumbs);
  pageTitle = title;
  updateBrowserTitle();
}

function updateBrowserTitle() {
  const parts = [
    pageContext,
    pageTitle,
    environmentName || DEFAULT_TITLE,
  ].filter(isPresent);

  document.title = Array.from(new Set(parts)).join(' - ');
}

function getLastBreadcrumbLabel(
  breadcrumbs: BrowserTitleBreadcrumbInput[] | string
) {
  const breadcrumbsArray = Array.isArray(breadcrumbs)
    ? breadcrumbs
    : [breadcrumbs];

  return (
    breadcrumbsArray
      .map((crumb) => (typeof crumb === 'string' ? crumb : crumb.label))
      .filter(isPresent)
      .at(-1) || undefined
  );
}

function isPresent(value: string | undefined): value is string {
  return !!value?.trim();
}

interface BrowserTitleBreadcrumb {
  label: string;
}

type BrowserTitleBreadcrumbInput = BrowserTitleBreadcrumb | string;
