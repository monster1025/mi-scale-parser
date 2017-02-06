package workers

import "github.com/currantlabs/gatt"

func getDescriptorByUUID(uuid gatt.UUID, descriptors []*gatt.Descriptor) *gatt.Descriptor {
	for _, descriptor := range descriptors {
		if !descriptor.UUID().Equal(uuid) {
			continue
		}
		return descriptor
	}
	return nil
}

func getServiceByUUID(uuid gatt.UUID, services []*gatt.Service) *gatt.Service {
	for _, service := range services {
		if !service.UUID().Equal(uuid) {
			continue
		}
		return service
	}
	return nil
}

func getCharByUUID(uid gatt.UUID, chars []*gatt.Characteristic) *gatt.Characteristic {
	for _, c := range chars {
		if !c.UUID().Equal(uid) {
			continue
		}
		return c
	}
	return nil
}
