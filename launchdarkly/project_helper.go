package launchdarkly

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	ldapi "github.com/launchdarkly/api-client-go"
)

func projectRead(d *schema.ResourceData, meta interface{}, isDataSource bool) error {
	client := meta.(*Client)
	projectKey := d.Get(KEY).(string)

	rawProject, res, err := handleRateLimit(func() (interface{}, *http.Response, error) {
		return client.ld.ProjectsApi.GetProject(client.ctx, projectKey)
	})
	// return nil error for resource reads but 404 for data source reads
	if isStatusNotFound(res) && !isDataSource {
		log.Printf("[WARN] failed to find project with key %q, removing from state if present", projectKey)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get project with key %q: %v", projectKey, err)
	}

	project := rawProject.(ldapi.Project)
	// the Id needs to be set on reads for the data source, but it will mess up the state for resource reads
	if isDataSource {
		d.SetId(project.Id)
	}
	_ = d.Set(KEY, project.Key)
	_ = d.Set(NAME, project.Name)

	envsRaw := environmentsToResourceData(project.Environments)
	err = d.Set(ENVIRONMENTS, envsRaw)
	if err != nil {
		return fmt.Errorf("could not set environments on project with key %q: %v", project.Key, err)
	}
	err = d.Set(TAGS, project.Tags)
	if err != nil {
		return fmt.Errorf("could not set tags on project with key %q: %v", project.Key, err)
	}
	err = d.Set(INCLUDE_IN_SNIPPET, project.IncludeInSnippetByDefault)
	if err != nil {
		return fmt.Errorf("could not set include_in_snippet on project with key %q: %v", project.Key, err)
	}
	return nil
}