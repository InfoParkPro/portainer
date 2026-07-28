type RefetchableQuery = {
  refetch: () => Promise<unknown>;
};

export function refetchSwarmStackResourceQueries({
  servicesQuery,
  tasksQuery,
  containersQuery,
}: {
  servicesQuery: RefetchableQuery;
  tasksQuery: RefetchableQuery;
  containersQuery?: RefetchableQuery;
}) {
  return Promise.all(
    [
      servicesQuery.refetch(),
      tasksQuery.refetch(),
      containersQuery?.refetch(),
    ].filter(Boolean)
  );
}
