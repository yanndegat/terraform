package openstack

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func resourceImageV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceImageV2Create,
		Read:   resourceImageV2Read,
		Update: resourceImageV2Update,
		Delete: resourceImageV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", ""),
			},

			"local_file_path": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image_source_url"},
			},

			"image_source_url": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"local_file_path"},
			},

			"image_cache_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  fmt.Sprintf("%s/.terraform/image_cache", os.Getenv("HOME")),
			},

			"visibility": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validateVisibility,
				Default:      images.ImageVisibilityPrivate,
			},

			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"container_format": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateContainerFormat,
			},

			"disk_format": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDiskFormat,
			},

			"min_disk_gb": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validatePositiveInt,
				Default:      0,
			},

			"min_ram_mb": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validatePositiveInt,
				Default:      0,
			},

			"protected": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"owner": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"checksum": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"size_bytes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"metadata": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},

			"created_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_update": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"file": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"schema": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateVisibility(v interface{}, k string) (ws []string, errors []error) {
	value := v.(images.ImageVisibility)
	if value == images.ImageVisibilityPublic || value == images.ImageVisibilityPrivate || value == images.ImageVisibilityShared || value == images.ImageVisibilityCommunity {
		return
	}
	errors = append(errors, fmt.Errorf("%q must be one of %q, %q, %q, %q", k, images.ImageVisibilityPublic, images.ImageVisibilityPrivate, images.ImageVisibilityCommunity, images.ImageVisibilityShared))
	return
}

func validatePositiveInt(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value > 0 {
		return
	}
	errors = append(errors, fmt.Errorf("%q must be a positive integer", k))
	return
}

var DiskFormats = [9]string{"ami", "ari", "aki", "vhd", "vmdk", "raw", "qcow2", "vdi", "iso"}

func validateDiskFormat(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	for i := range DiskFormats {
		if value == DiskFormats[i] {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%q must be one of %v", k, DiskFormats))
	return
}

var ContainerFormats = [9]string{"ami", "ari", "aki", "bare", "ovf"}

func validateContainerFormat(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	for i := range ContainerFormats {
		if value == ContainerFormats[i] {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%q must be one of %v", k, ContainerFormats))
	return
}

func imageV2VisibitlityFromString(v string) (images.ImageVisibility, error) {
	if images.ImageVisibility(v) == images.ImageVisibilityCommunity {
		return images.ImageVisibilityCommunity, nil
	}
	if images.ImageVisibility(v) == images.ImageVisibilityPrivate {
		return images.ImageVisibilityPrivate, nil
	}
	if images.ImageVisibility(v) == images.ImageVisibilityPublic {
		return images.ImageVisibilityPublic, nil
	}
	if images.ImageVisibility(v) == images.ImageVisibilityShared {
		return images.ImageVisibilityShared, nil
	}

	return "", fmt.Errorf("Error unkown image visibility %q", v)
}

func resourceImageV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	imageClient, err := config.imageV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack image client: %s", err)
	}

	createOpts := &images.CreateOpts{
		Name:             d.Get("name").(string),
		ID:               d.Get("id").(string),
		ContainerFormat:  d.Get("container_format").(string),
		DiskFormat:       d.Get("disk_format").(string),
		MinDiskGigabytes: d.Get("min_disk_gb").(int),
		MinRAMMegabytes:  d.Get("min_ram_mb").(int),
		Protected:        d.Get("protected").(bool),
	}

	if v := d.Get("visibility").(string); v != "" {
		vis, err := imageV2VisibitlityFromString(v)
		if err != nil {
			return err
		}
		createOpts.Visibility = &vis
	}

	if tags := d.Get("tags"); tags != nil {
		ts := tags.([]interface{})
		createOpts.Tags = make([]string, len(ts))
		for _, v := range ts {
			createOpts.Tags = append(createOpts.Tags, v.(string))
		}
	}

	if props := d.Get("properties"); props != nil {
		ps := props.(map[string]interface{})
		createOpts.Properties = make(map[string]string)
		for k, v := range ps {
			createOpts.Properties[k] = v.(string)
		}
	}

	d.Partial(true)

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	newImg, err := images.Create(imageClient, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creating Image: %s", err)
	}

	d.SetId(newImg.ID)

	// downloading/getting image file props
	imgFilePath, err := resourceImageFile(d)
	if err != nil {
		return fmt.Errorf("Error opening file for Image: %s", err)

	}
	fileSize, fileChecksum, err := resourceImageFileProps(imgFilePath)
	if err != nil {
		return fmt.Errorf("Error getting file props: %s", err)
	}

	// upload
	imgFile, err := os.Open(imgFilePath)
	if err != nil {
		return fmt.Errorf("Error opening file %q: %s", imgFilePath, err)
	}
	defer imgFile.Close()
	log.Printf("[WARN] Uploading image %s (%d bytes). This can be pretty long.", d.Id(), fileSize)

	res := images.Upload(imageClient, d.Id(), imgFile)
	if res.Err != nil {
		return fmt.Errorf("Error while uploading file %q: %s", imgFilePath, res.Err)
	}

	//wait for active
	stateConf := &resource.StateChangeConf{
		Pending:    []string{string(images.ImageStatusQueued), string(images.ImageStatusSaving)},
		Target:     []string{string(images.ImageStatusActive)},
		Refresh:    ImageV2RefreshFunc(imageClient, d.Id(), fileSize, fileChecksum),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Image: %s", err)
	}

	d.Partial(false)

	return resourceImageV2Read(d, meta)
}

func fileMD5Checksum(f *os.File) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func resourceImageFileProps(filename string) (int64, string, error) {
	var filesize int64
	var filechecksum string

	file, err := os.Open(filename)
	if err != nil {
		return -1, "", fmt.Errorf("Error opening file for Image: %s", err)

	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return -1, "", fmt.Errorf("Error reading image file %q: %s", file.Name(), err)
	}

	filesize = fstat.Size()
	filechecksum, err = fileMD5Checksum(file)

	if err != nil {
		return -1, "", fmt.Errorf("Error computing image file %q checksum: %s", file.Name(), err)
	}

	return filesize, filechecksum, nil
}

func resourceImageFile(d *schema.ResourceData) (string, error) {
	if filename := d.Get("local_file_path").(string); filename != "" {
		return filename, nil
	} else if furl := d.Get("image_source_url").(string); furl != "" {
		dir := d.Get("image_cache_path").(string)
		os.MkdirAll(dir, 0700)
		filename := filepath.Join(dir, fmt.Sprintf("%x.img", md5.Sum([]byte(furl))))

		if _, err := os.Stat(filename); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("Error while trying to access file %q: %s", filename, err)
			}
			log.Printf("[DEBUG] File doens't exists %s. will download from %s", filename, furl)
			file, err := os.Create(filename)
			if err != nil {
				return "", fmt.Errorf("Error creating file %q: %s", filename, err)
			}
			defer file.Close()
			resp, err := http.Get(furl)
			if err != nil {
				return "", fmt.Errorf("Error downloading image from %q", furl)
			}
			defer resp.Body.Close()

			if _, err = io.Copy(file, resp.Body); err != nil {
				return "", fmt.Errorf("Error downloading image %q to file %q: %s", furl, filename, err)
			}
			return filename, nil
		} else {
			log.Printf("[DEBUG] File exists %s", filename)
			return filename, nil
		}
	} else {
		return "", fmt.Errorf("Error in config. no file specified")
	}
}

func resourceImageV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	imageClient, err := config.imageV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack image client: %s", err)
	}

	img, err := images.Get(imageClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "image")
	}

	log.Printf("[DEBUG] Retrieved Image %s: %+v", d.Id(), img)

	d.Set("owner", img.Owner)
	d.Set("status", img.Status)
	d.Set("file", img.File)
	d.Set("schema", img.Schema)
	d.Set("checksum", img.Checksum)
	d.Set("size_bytes", img.SizeBytes)
	d.Set("metadata", img.Metadata)
	d.Set("created_date", img.CreatedDate)
	d.Set("last_update", img.LastUpdate)
	d.Set("id", img.ID)

	return nil
}

func resourceImageV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	imageClient, err := config.imageV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack image client: %s", err)
	}

	updateOpts := make(images.UpdateOpts, 0)

	if d.HasChange("visibility") {
		v := images.UpdateVisibility{Visibility: d.Get("visibility").(images.ImageVisibility)}
		updateOpts = append(updateOpts, v)
	}

	if d.HasChange("name") {
		v := images.ReplaceImageName{NewName: d.Get("name").(string)}
		updateOpts = append(updateOpts, v)
	}

	if d.HasChange("tags") {
		v := images.ReplaceImageTags{NewTags: d.Get("tags").([]string)}
		updateOpts = append(updateOpts, v)
	}

	log.Printf("[DEBUG] Update Options: %#v", updateOpts)

	_, err = images.Update(imageClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error updating image: %s", err)
	}

	return resourceImageV2Read(d, meta)
}

func resourceImageV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	imageClient, err := config.imageV2Client(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack image client: %s", err)
	}

	log.Printf("[DEBUG] Deleting Image %s", d.Id())
	if err := images.Delete(imageClient, d.Id()).Err; err != nil {
		return fmt.Errorf("Error deleting Image: %s", err)
	}

	d.SetId("")
	return nil
}

func ImageV2RefreshFunc(client *gophercloud.ServiceClient, id string, fileSize int64, checksum string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		img, err := images.Get(client, id).Extract()
		if err != nil {
			return nil, "", err
		}
		log.Printf("[DEBUG] OpenStack image status is: %s", img.Status)

		if img.Checksum != checksum || int64(img.SizeBytes) != fileSize {
			return img, fmt.Sprintf("%s", img.Status), fmt.Errorf("Error wrong size %v or checksum %q", img.SizeBytes, img.Checksum)
		}

		return img, fmt.Sprintf("%s", img.Status), nil
	}
}
